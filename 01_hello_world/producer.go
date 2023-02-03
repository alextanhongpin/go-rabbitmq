package main

import (
	"context"
	"log"
	"time"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// 1. Dial rabbitmq.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Panicf("failed to dial rabbitmq: %v", err)
	}
	defer conn.Close()

	// 2. Create a channel.
	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("failed to create channel: %v", err)
	}
	defer ch.Close()

	// 3. Declare a queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		QueueName: "hello",
	})
	if err != nil {
		log.Panicf("failed to declare queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 4. Publish message to the queue.
	body := "hello world"
	err = amqplib.PublishWithContext(ctx, ch, amqplib.PublishWithContextOption{
		RoutingKey: q.Name,
		Message: amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	})
	if err != nil {
		log.Panicf("failed to publish: %v", err)
	}

	log.Printf("[x] Sent %s\n", body)
}
