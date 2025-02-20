package repository

import (
	"context"
	"time"

	"github.com/shrey258/task_management/internal/database"
	"github.com/shrey258/task_management/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		collection: database.GetDB().Collection("tasks"),
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Status == "" {
		task.Status = models.StatusTodo
	}

	result, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TaskRepository) Update(ctx context.Context, id primitive.ObjectID, update *models.TaskUpdate) error {
	updateDoc := bson.M{"updated_at": time.Now()}

	if update.Title != nil {
		updateDoc["title"] = *update.Title
	}
	if update.Description != nil {
		updateDoc["description"] = *update.Description
	}
	if update.Priority != nil {
		updateDoc["priority"] = *update.Priority
	}
	if update.Status != nil {
		updateDoc["status"] = *update.Status
	}
	if update.DueDate != nil {
		updateDoc["due_date"] = *update.DueDate
	}
	if update.AssignedTo != nil {
		updateDoc["assigned_to"] = *update.AssignedTo
	}
	if update.Tags != nil {
		updateDoc["tags"] = *update.Tags
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateDoc},
	)
	return err
}

func (r *TaskRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Task, error) {
	var task models.Task
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Find(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
	filterDoc := bson.M{}

	if filter.Status != nil {
		filterDoc["status"] = *filter.Status
	}
	if filter.Priority != nil {
		filterDoc["priority"] = *filter.Priority
	}
	if filter.AssignedTo != nil {
		filterDoc["assigned_to"] = *filter.AssignedTo
	}
	if filter.CreatedBy != nil {
		filterDoc["created_by"] = *filter.CreatedBy
	}
	if len(filter.Tags) > 0 {
		filterDoc["tags"] = bson.M{"$in": filter.Tags}
	}
	if filter.DueBefore != nil {
		filterDoc["due_date"] = bson.M{"$lt": *filter.DueBefore}
	}
	if filter.DueAfter != nil {
		if _, exists := filterDoc["due_date"]; exists {
			filterDoc["due_date"].(bson.M)["$gt"] = *filter.DueAfter
		} else {
			filterDoc["due_date"] = bson.M{"$gt": *filter.DueAfter}
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filterDoc, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
