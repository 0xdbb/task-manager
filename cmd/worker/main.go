package main

import (
	"fmt"
	"log"
	"task-manager/config"
	"task-manager/internal/queue"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO: Listen to queue

func main() {

	// -------Load Config-------
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	// -------Initialize QueueManager-------
	newQm, err := queue.NewQueueManager(config.RMQ_ADDRESS)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	defer newQm.Close()

	// -------Declare Queue-------
	q, err := newQm.DeclareQueue(queue.QueueOptions{
		Name:       "task_queue",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       amqp.Table{"x-max-priority": 10}, // Enable priority queue
	})

	if err != nil {
		log.Fatalf("queue error: %s", err)
	}

	// Apply QoS (Prefetch Count)
	err = newQm.SetQos(
		1,
		0,
		false,
	)

	if err != nil {
		newQm.Close()
		log.Fatalf("failed to set Qos: %s", err)
	}

	taskMsg, err := newQm.Consume(queue.ConsumerOptions{
		QueueName: q.Name,
		WorkerTag: "worker-1",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	})

	var forever chan struct{}

	go func() {
		for t := range taskMsg {
			processTask(t.Body)
			log.Printf("Done")
			t.Ack(false)
		}

	}()

	log.Printf(" [*] Waiting for tasks. To exit press CTRL+C")
	<-forever

}

func processTask(t []byte) {
	log.Printf("Received a message: %s", t)
}
