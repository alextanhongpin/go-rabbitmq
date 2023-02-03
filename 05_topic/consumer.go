package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

const exchangeName = "logs_topic"

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %v", msg, err)
	}
}

func main() {
	var topic string
	flag.StringVar(&topic, "topic", "*.info", "the logs topic, one of <pattern>.<info|warning|error>")
	flag.Parse()

	switch {
	case strings.Contains(topic, ".info"):
	case strings.Contains(topic, ".warning"):
	case strings.Contains(topic, ".error"):
	default:
		panic("-topic must adhere to the format <facility>.<info|warning|error>")
	}

	// 1. Dial amqp.
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	failOnError(err, "failed to dial")
	defer conn.Close()

	// 2. Create a channel.
	ch, err := conn.Channel()
	failOnError(err, "failed to create channel")
	defer ch.Close()

	// 3. Declare an exclusive queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		Exclusive: true,
	})
	failOnError(err, "failed to declare queue")

	// 4. Bind the queue to the exchange with the topic pattern.
	err = amqplib.QueueBind(ch, amqplib.QueueBindOption{
		QueueName:    q.Name,
		BindingKey:   topic,
		ExchangeName: exchangeName,
	})
	failOnError(err, "failed to bind queue")

	// 5. Create a consumer with auto acknowledgement.
	msgs, err := amqplib.Consume(ch, amqplib.ConsumeOption{
		QueueName: q.Name,
		AutoAck:   true,
	})
	failOnError(err, "failed to create consumer")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case msg := <-msgs:
				log.Printf("[x] Received %s from %s\n", msg.Body, msg.RoutingKey)
			case <-done:
				return
			}
		}
	}()

	log.Printf("[*] listening to topic exchange [%s]. Press ctrl + c to cancel\n", topic)
	wg.Wait()
}
