package main

import (
	"context"
	"flag"
	"log"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var body, queue string
	flag.StringVar(&body, "body", "", "body to send")
	flag.StringVar(&queue, "queue", "hello", "queue name")
	flag.Parse()
	if body == "" {
		panic("-body is required")
	}

	// 1. Dial amqp.
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	if err != nil {
		log.Panicf("failed to dial: %v", err)
	}
	defer conn.Close()

	// 2. Create a channel.
	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("failed to create channel: %v", err)
	}
	defer ch.Close()

	// 3. Declare a new durable queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		QueueName: queue,
		Durable:   true,
	})
	if err != nil {
		log.Panicf("failed to declare queue: %v", err)
	}

	// 4. Publish persisten message to the queue.
	ctx := context.Background()
	err = amqplib.PublishWithContext(ctx, ch, amqplib.PublishWithContextOption{
		RoutingKey: q.Name,
		Message: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Type:         "text/plain",
			Body:         []byte(body),
		},
	})
	if err != nil {
		log.Panicf("failed to publish: %v", err)
	}

	log.Printf("[x] Sent %s", body)
}
