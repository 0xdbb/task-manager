package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	db "task-manager/internal/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
)

var InProgress = "IN_PROGRESS"

// TaskWorker processes tasks and updates the database.
type TaskWorker struct {
	Task    amqp.Delivery
	Service *db.Service
}

// Task execution logic
func (tw *TaskWorker) DoTask(ctx context.Context) {
	id := string(tw.Task.MessageId) // Assuming MessageId stores task UUID
	log.Printf("Processing task ID: %s", id)

	taskID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Failed to update task status to IN_PROGRESS: %v", err)
		return
	}

	args := db.UpdateTaskStatusParams{
		ID: taskID,
		Status: db.NullTaskStatus{
			TaskStatus: db.TaskStatus(InProgress),
			Valid:      true,
		},
	}

	// Update task status to "IN_PROGRESS"
	_, err = tw.Service.UpdateTaskStatus(context.Background(), args)
	if err != nil {
		log.Printf("Failed to update task status to IN_PROGRESS: %v", err)
		return
	}

	// Task Processing
	_ = processTask(tw.Task.Body)

	// TODO: Check update if Task processing doesn't fail

	// Update task status to "COMPLETED" and store result
	args = db.UpdateTaskStatusParams{
		ID: taskID,
		Status: db.NullTaskStatus{
			TaskStatus: db.TaskStatus(InProgress),
			Valid:      true,
		},
		Result: pgtype.Text{
			String: "Result",
			Valid:  true,
		},
	}

	// Acknowledge the message after successful processing
	tw.Task.Ack(false)
}

// processTask simulates task execution (modify this for real processing)
func processTask(task []byte) string {
	log.Printf("Processing task data: %s", task)
	time.Sleep(2 * time.Second) // Simulate work delay
	return fmt.Sprintf("Processed: %s", task)
}
