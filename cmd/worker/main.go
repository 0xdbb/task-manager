package main

// TODO: write tests
// TODO: call goose in code
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
	"task-manager/internal/weather"
	"task-manager/internal/worker"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TaskProcessor implements the worker.TaskProcessor interface

func main() {
	// ------- Load Config -------
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	// ------- Initialize Database Service -------
	dbURL := config.DbUrlDev
	if config.Production == "1" {
		dbURL = config.DbUrl
	}
	dbService := db.NewService(dbURL)

	// ------- Initialize QueueManager -------
	qm, err := queue.NewQueueManager(config.RMQAddress)
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
	weatherProcessor := weather.NewWeatherProcessor(config.WeatherApiKey)
	worker := worker.New(dbService, weatherProcessor, 30*time.Minute)
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
