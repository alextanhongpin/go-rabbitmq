package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "logs"

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %v", msg, err)
	}
}

func main() {
	var body string
	flag.StringVar(&body, "body", "", "the body to send")
	flag.Parse()

	if body == "" {
		log.Panic("-body is required")
	}

	// 1. Dial amqp.
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	failOnError(err, "failed to dial")
	defer conn.Close()

	// 2. Create channel.
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")
	defer ch.Close()

	// 3. Create a new durable exchange with type fanout.
	err = amqplib.ExchangeDeclare(ch, amqplib.ExchangeDeclareOption{
		ExchangeName: exchange,
		Type:         amqplib.ExchangeTypeFanout,
		Durable:      true,
	})
	failOnError(err, "failed to declare an exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 4. Publish message.
	err = amqplib.PublishWithContext(ctx, ch, amqplib.PublishWithContextOption{
		ExchangeName: exchange,
		Message: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		},
	})
	if err != nil {
		log.Panicf("failed to publish: %v", err)
	}

	log.Printf("[x] Sent %s", body)
}
