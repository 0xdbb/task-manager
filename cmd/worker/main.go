package main

// TODO: write tests
// TODO: deployment with docker and kubernetes

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-manager/internal/config"
	db "task-manager/internal/database/sqlc"
	"task-manager/internal/queue"
	"task-manager/internal/worker"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TaskProcessor implements the worker.TaskProcessor interface
type TaskProcessor struct {
	service *db.Service
}

func (p *TaskProcessor) ProcessTask(body []byte) (string, error) {
	// Implement your actual task processing logic here
	// For example:
	// 1. Parse the task payload
	// 2. Perform the required operations (send email, generate report, etc.)
	// 3. Return result or error

	// Currently just logging and returning a success message
	// log.Printf("Processing task: %s", string(body))
	return "Task processed successfully", nil
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
	dbService := db.NewService(dbURL)

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

	// ------- Initialize Worker -------
	taskProcessor := &TaskProcessor{service: dbService}
	worker := worker.New(dbService, taskProcessor, 30*time.Minute)
	worker.Consume(taskMsgs)

	// ------- Graceful Shutdown Setup -------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ------- Start Worker in Goroutine -------
	workerDone := make(chan error, 1)
	go func() {
		log.Println("Worker started")
		workerDone <- worker.Start()
	}()

	// ------- Wait for Shutdown Signal -------
	select {
	case sig := <-quit:
		log.Printf("Received signal: %v. Shutting down...", sig)
		// Additional cleanup can be done here if needed
		return
	case err := <-workerDone:
		if err != nil {
			log.Printf("Worker stopped with error: %v", err)
		} else {
			log.Println("Worker stopped normally")
		}
	}
}
