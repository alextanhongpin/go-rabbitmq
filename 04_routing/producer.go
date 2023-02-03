package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

const exchangeName = "logs_direct"

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %v", msg, err)
	}
}

func main() {
	var body, severity string
	flag.StringVar(&body, "body", "", "body to send")
	flag.StringVar(&severity, "severity", "info", "the logs severity, one of info, warning or error")
	flag.Parse()

	if body == "" {
		log.Panic("-body is required")
	}
	switch severity {
	case "info", "warning", "error":
	default:
		panic("-severity must be one of info, warning, or error")
	}

	// 1. Dial amqp.
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	failOnError(err, "failed to dial")
	defer conn.Close()

	// 2. Create a channel.
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")
	defer ch.Close()

	// 3. Declare a durable exchange with type direct.
	err = amqplib.ExchangeDeclare(ch, amqplib.ExchangeDeclareOption{
		ExchangeName: exchangeName,
		Type:         amqplib.ExchangeTypeDirect,
		Durable:      true,
	})
	failOnError(err, "failed to create exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 4. Publish message to the exchange.
	err = amqplib.PublishWithContext(ctx, ch, amqplib.PublishWithContextOption{
		ExchangeName: exchangeName,
		RoutingKey:   severity,
		Message: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Type:         "text/plain",
			Body:         []byte(body),
		},
	})
	failOnError(err, "failed to publish")

	log.Printf("[x] Sent %s (%s)", body, severity)
}
