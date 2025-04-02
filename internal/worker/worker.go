package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	db "task-manager/internal/database/sqlc"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TaskProcessor interface {
	ProcessTask(body []byte) (string, error)
}

type Worker struct {
	interrupt  chan os.Signal
	complete   chan error
	timeout    <-chan time.Time
	tasksChan  <-chan amqp.Delivery
	service    *db.Service
	processor  TaskProcessor
	maxRetries int
	logger     *log.Logger
}

var (
	ErrTimeout   = errors.New("received timeout")
	ErrInterrupt = errors.New("received interrupt")
)

func New(service *db.Service, processor TaskProcessor, timeout time.Duration) *Worker {
	return &Worker{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(timeout),
		service:   service,
		processor: processor,
		logger:    log.New(os.Stdout, "worker: ", log.LstdFlags),
	}
}

func (w *Worker) Consume(taskMsgChan <-chan amqp.Delivery) {
	w.tasksChan = taskMsgChan
}

func (w *Worker) Start() error {
	signal.Notify(w.interrupt, os.Interrupt)

	go func() {
		w.complete <- w.run()
	}()

	select {
	case err := <-w.complete:
		return err
	case <-w.timeout:
		return ErrTimeout
	}
}

func (w *Worker) run() error {
	for task := range w.tasksChan {
		if w.gotInterrupt() {
			return ErrInterrupt
		}

		if err := w.processTask(task); err != nil {
			w.logger.Printf("Task processing failed: %v", err)
			continue
		}
	}
	return nil
}

func (w *Worker) processTask(task amqp.Delivery) error {
	// Parse task from message
	t, err := parseJson[db.Task](task.Body)
	if err != nil {
		w.logger.Printf("Failed to parse task: %v", err)
		task.Nack(false, false) // Discard malformed message
		return err
	}

	w.logger.Printf("Processing task ID: %s", t.ID)

	// Update status to IN_PROGRESS
	if err := w.updateTaskStatus(t.ID, db.TaskStatusINPROGRESS, ""); err != nil {
		w.logger.Printf("Failed to update task status: %v", err)
		return err
	}

	// Process the task
	result, err := w.processor.ProcessTask([]byte(t.Payload))
	if err != nil {
		w.logger.Printf("Task processing failed: %v", err)
		// Update status to FAILED
		if updateErr := w.updateTaskStatus(t.ID, db.TaskStatusFAILED, err.Error()); updateErr != nil {
			w.logger.Printf("Failed to update failed task status: %v", updateErr)
		}
		return err
	}

	// Update status to COMPLETED with result
	if err := w.updateTaskStatus(t.ID, db.TaskStatusCOMPLETED, result); err != nil {
		w.logger.Printf("Failed to update completed task status: %v", err)
		task.Nack(false, w.shouldRetry(err))
		return err
	}

	// Acknowledge successful processing
	task.Ack(false)
	w.logger.Printf("Successfully processed task ID: %s", t.ID)
	return nil
}

// Prevents infinite retries for non-recoverable errors
func (w *Worker) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Avoid retrying for invalid enum values (SQLSTATE 22P02)
	if strings.Contains(err.Error(), "SQLSTATE 22P02") {
		return false
	}

	return true
}

func (w *Worker) updateTaskStatus(id uuid.UUID, status db.TaskStatus, result string) error {
	args := db.UpdateTaskStatusParams{
		ID: id,
		Status: db.NullTaskStatus{
			TaskStatus: status,
			Valid:      true,
		},
	}

	if result != "" {
		args.Result = pgtype.Text{
			String: result,
			Valid:  true,
		}
	}

	_, err := w.service.UpdateTaskStatus(context.Background(), args)
	return err
}

func (w *Worker) gotInterrupt() bool {
	select {
	case <-w.interrupt:
		signal.Stop(w.interrupt)
		return true
	default:
		return false
	}
}

// parseJson is a generic function that unmarshals JSON data into a specified type.
// It provides type-safe parsing with detailed error handling.
func parseJson[T any](typeBytes []byte) (T, error) {
	var zero T // Get the zero value of type T

	// Early return for empty input
	if len(typeBytes) == 0 {
		return zero, fmt.Errorf("empty task data")
	}

	// Create a new instance of type T
	var task T

	// Unmarshal with strict decoding
	decoder := json.NewDecoder(bytes.NewReader(typeBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&task); err != nil {
		// Enhanced error handling
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return zero, fmt.Errorf("invalid JSON syntax at position %d: %w", syntaxErr.Offset, err)
		case errors.As(err, &unmarshalErr):
			return zero, fmt.Errorf("invalid type for field %q (expected %s, got %s): %w",
				unmarshalErr.Field,
				unmarshalErr.Type,
				unmarshalErr.Value,
				err)
		default:
			return zero, fmt.Errorf("failed to parse task: %w", err)
		}
	}

	// Optional: Add validation if T implements a Validator interface
	if validator, ok := any(task).(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return zero, fmt.Errorf("task validation failed: %w", err)
		}
	}

	return task, nil
}
