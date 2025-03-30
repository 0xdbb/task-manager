package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-manager/internal/config"
	"task-manager/internal/queue"
	"task-manager/internal/runner"
	"task-manager/internal/worker"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Worker struct for processing tasks
type TaskWorker struct {
	task amqp.Delivery
}

// Implement Worker interface for TaskWorker
func (tw *TaskWorker) Task() {
	processTask(tw.task.Body)

	// Acknowledge only after processing is complete
	tw.task.Ack(false)
}

func main() {
	// -------Load Config-------
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	// -------Initialize QueueManager-------
	qm, err := queue.NewQueueManager(config.RMQ_ADDRESS)
	if err != nil {
		log.Fatalf("Queue error: %s", err)
	}
	defer qm.Close()

	// -------Declare Queue-------
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

	// -------Apply QoS (Prefetch Count)-------
	err = qm.SetQos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %s", err)
	}

	// -------Consume Messages-------
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

	// -------Initialize Worker Pool-------
	workerPool := worker.New(5) // Create a pool with 5 workers

	// -------Initialize Runner for Graceful Shutdown-------
	r := runner.NewRunner(time.Minute * 30)

	// Capture termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Add tasks to the runner
	r.Add(func(int) {
		for t := range taskMsgs {
			workerPool.Run(&TaskWorker{task: t})
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

// Process each task (Modify this to your actual task logic)
func processTask(task []byte) {
	log.Printf("Processing task: %s", task)
}
