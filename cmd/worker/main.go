package main
// TODO: update status
// TODO: Polling for task status
// TODO: Rate limiting and Throttling
// TODO: write tests
// TODO: deployment with docker and kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-manager/internal/config"
	db "task-manager/internal/database/sqlc"
	"task-manager/internal/queue"
	"task-manager/internal/runner"
	"task-manager/internal/worker"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TaskWorker processes tasks and updates the database.
type TaskWorker struct {
	task    amqp.Delivery
	service *db.Service
}

// Task execution logic
func (tw *TaskWorker) Task(ctx context.Context) {
	taskID := string(tw.task.MessageId) // Assuming MessageId stores task UUID
	log.Printf("Processing task ID: %s", taskID)

	// Update task status to "IN_PROGRESS"
	// err := tw.service.UpdateTaskStatus(context.Background(), db.UpdateTaskStatusParams{
	// 	ID:     taskID,
	// 	Status: "IN_PROGRESS",
	// })
	// if err != nil {
	// 	log.Printf("Failed to update task status to IN_PROGRESS: %v", err)
	// 	return
	// }

	// Simulate task processing
	_ = processTask(tw.task.Body)

	// Update task status to "COMPLETED" and store result
	// _, err = tw.service.UpdateTaskStatus(context.Background(), db.UpdateTaskStatusParams{
	// 	ID:     ,
	// 	Status: db.NullTaskStatus{},
	// 	Result: pgtype.Text{},
	// })
	// if err != nil {
	// 	log.Printf("Failed to update task status to COMPLETED: %v", err)
	// 	return
	// }

	// Acknowledge the message after successful processing
	tw.task.Ack(false)
}

// processTask simulates task execution (modify this for real processing)
func processTask(task []byte) string {
	log.Printf("Processing task data: %s", task)
	time.Sleep(2 * time.Second) // Simulate work delay
	return fmt.Sprintf("Processed: %s", task)
}

func main() {
	// ------- Load Config -------
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	// ------- Initialize Database Service -------
	dbURL := config.DB_URL_DEV
	if config.PRODUCTION == "1" {
		dbURL = config.DB_URL
	}
	newService := db.NewService(dbURL)

	// ------- Initialize QueueManager -------
	qm, err := queue.NewQueueManager(config.RMQ_ADDRESS)
	if err != nil {
		log.Fatalf("Queue error: %s", err)
	}
	defer qm.Close()

	// ------- Declare Queue -------
	q, err := qm.DeclareQueue(queue.QueueOptions{
		Name:       "task_queue",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       amqp.Table{"x-max-priority": 10},
	})
	if err != nil {
		log.Fatalf("Queue declaration error: %s", err)
	}

	// ------- Apply QoS -------
	err = qm.SetQos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %s", err)
	}

	// ------- Consume Messages -------
	taskMsgs, err := qm.Consume(queue.ConsumerOptions{
		QueueName: q.Name,
		WorkerTag: "worker-1",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})
	if err != nil {
		log.Fatalf("Failed to consume messages: %s", err)
	}

	// ------- Initialize Worker Pool -------
	workerPool := worker.New(5, 30*time.Minute) // Pool with 5 sub-workers

	// ------- Initialize Runner for Graceful Shutdown -------
	r := runner.New()

	// Capture termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Add tasks to the runner
	r.Add(func(int) {
		for t := range taskMsgs {
			workerPool.Run(&TaskWorker{task: t, service: newService})
		}
	})

	// Start the runner in a separate goroutine
	go func() {
		if err := r.Start(); err != nil {
			log.Println("Runner stopped:", err)
		}
	}()

	log.Println(" [*] Waiting for tasks. Press CTRL+C to exit.")

	// Block until a termination signal is received
	<-stop

	// Shutdown gracefully
	log.Println("Shutting down...")
	workerPool.Shutdown()
	qm.Close()
	log.Println("Shutdown complete.")
}

