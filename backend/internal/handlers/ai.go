package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shrey258/task_management/internal/ai"
)

type AIHandler struct {
	gemini *ai.GeminiService
}

func NewAIHandler(gemini *ai.GeminiService) *AIHandler {
	return &AIHandler{
		gemini: gemini,
	}
}

type GenerateTaskSuggestionsRequest struct {
	Description string `json:"description"`
}

type AnalyzeTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *AIHandler) GenerateTaskSuggestions(c *fiber.Ctx) error {
	var req GenerateTaskSuggestionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "description is required",
		})
	}

	suggestions, err := h.gemini.GenerateTaskSuggestions(c.Context(), req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate suggestions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"suggestions": suggestions,
	})
}

func (h *AIHandler) AnalyzeTask(c *fiber.Ctx) error {
	var req AnalyzeTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Title == "" || req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "title and description are required",
		})
	}

	suggestion, err := h.gemini.AnalyzeTaskPriority(c.Context(), req.Title, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to analyze task",
		})
	}

	return c.Status(fiber.StatusOK).JSON(suggestion)
}
