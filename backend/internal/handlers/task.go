package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shrey258/task_management/internal/models"
	"github.com/shrey258/task_management/internal/repository"
	"github.com/shrey258/task_management/internal/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	taskRepo *repository.TaskRepository
	hub      *websocket.Hub
}

func NewTaskHandler(taskRepo *repository.TaskRepository, hub *websocket.Hub) *TaskHandler {
	return &TaskHandler{
		taskRepo: taskRepo,
		hub:      hub,
	}
}

type CreateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	DueDate     time.Time `json:"due_date"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
	Tags        []string  `json:"tags"`
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	userID, ok := c.Locals("user_id").(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var assignedToID *primitive.ObjectID
	if req.AssignedTo != "" {
		id, err := primitive.ObjectIDFromHex(req.AssignedTo)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid assigned_to id",
			})
		}
		assignedToID = &id
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    models.TaskPriority(req.Priority),
		Status:      models.StatusTodo,
		DueDate:     req.DueDate,
		CreatedBy:   userID,
		AssignedTo:  assignedToID,
		Tags:        req.Tags,
	}

	if err := h.taskRepo.Create(c.Context(), task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create task",
		})
	}

	// Notify relevant users about the new task
	h.hub.BroadcastToAll(websocket.Message{
		Type:    "task_created",
		Payload: task,
	})

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	var update models.TaskUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.taskRepo.Update(c.Context(), taskID, &update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update task",
		})
	}

	// Fetch updated task
	task, err := h.taskRepo.FindByID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch updated task",
		})
	}

	// Notify about task update
	h.hub.BroadcastToAll(websocket.Message{
		Type:    "task_updated",
		Payload: task,
	})

	return c.Status(fiber.StatusOK).JSON(task)
}

func (h *TaskHandler) GetTasks(c *fiber.Ctx) error {
	var filter models.TaskFilter

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		taskStatus := models.TaskStatus(status)
		filter.Status = &taskStatus
	}
	if priority := c.Query("priority"); priority != "" {
		taskPriority := models.TaskPriority(priority)
		filter.Priority = &taskPriority
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		id, err := primitive.ObjectIDFromHex(assignedTo)
		if err == nil {
			filter.AssignedTo = &id
		}
	}

	tasks, err := h.taskRepo.Find(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch tasks",
		})
	}

	return c.Status(fiber.StatusOK).JSON(tasks)
}

func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	task, err := h.taskRepo.FindByID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch task",
		})
	}
	if task == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(task)
}

func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	if err := h.taskRepo.Delete(c.Context(), taskID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete task",
		})
	}

	// Notify about task deletion
	h.hub.BroadcastToAll(websocket.Message{
		Type:    "task_deleted",
		Payload: taskID,
	})

	return c.Status(fiber.StatusNoContent).Send(nil)
}
