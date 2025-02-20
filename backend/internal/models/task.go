package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskPriority string
type TaskStatus string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"

	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
)

type Task struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Title       string              `json:"title" bson:"title"`
	Description string              `json:"description" bson:"description"`
	Priority    TaskPriority        `json:"priority" bson:"priority"`
	Status      TaskStatus          `json:"status" bson:"status"`
	DueDate     time.Time           `json:"due_date" bson:"due_date"`
	CreatedBy   primitive.ObjectID  `json:"created_by" bson:"created_by"`
	AssignedTo  *primitive.ObjectID `json:"assigned_to,omitempty" bson:"assigned_to,omitempty"`
	Tags        []string            `json:"tags" bson:"tags"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
}

type TaskUpdate struct {
	Title       *string              `json:"title,omitempty"`
	Description *string              `json:"description,omitempty"`
	Priority    *TaskPriority        `json:"priority,omitempty"`
	Status      *TaskStatus          `json:"status,omitempty"`
	DueDate     *time.Time           `json:"due_date,omitempty"`
	AssignedTo  *primitive.ObjectID  `json:"assigned_to,omitempty"`
	Tags        *[]string            `json:"tags,omitempty"`
}

type TaskFilter struct {
	Status     *TaskStatus          `json:"status,omitempty"`
	Priority   *TaskPriority        `json:"priority,omitempty"`
	AssignedTo *primitive.ObjectID  `json:"assigned_to,omitempty"`
	CreatedBy  *primitive.ObjectID  `json:"created_by,omitempty"`
	Tags       []string             `json:"tags,omitempty"`
	DueBefore  *time.Time           `json:"due_before,omitempty"`
	DueAfter   *time.Time           `json:"due_after,omitempty"`
}
