package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	var severity string
	flag.StringVar(&severity, "severity", "info", "the logs severity, one of info, warning or error")
	flag.Parse()

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

	// 3. Declare an exclusive queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		Exclusive: true,
	})
	failOnError(err, "failed to declare queue")

	// 4. Bind the queue to the exchange.
	err = amqplib.QueueBind(ch, amqplib.QueueBindOption{
		QueueName:    q.Name,
		BindingKey:   severity,
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
				log.Printf("[x] Received %s\n", msg.Body)
			case <-done:
				return
			}
		}
	}()

	log.Printf("[*] listening to direct exchange [%s]. Press ctrl + c to cancel\n", severity)
	wg.Wait()
}
