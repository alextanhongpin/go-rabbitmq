package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	// 1. Dial amqp.
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672")
	failOnError(err, "failed to dial")
	defer conn.Close()

	// 2. Create a channel.
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

	// 4. Declare an exclusive queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		Exclusive: true,
	})
	failOnError(err, "failed to declare queue")

	// 5. Bind the exchange to the queue.
	err = amqplib.QueueBind(ch, amqplib.QueueBindOption{
		QueueName:    q.Name,
		ExchangeName: exchange,
	})
	failOnError(err, "failed to bind queue")

	// 6. Create a consumer for the queue with auto acknowledgement.
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
				fmt.Printf("[x] %s\n", msg.Body)
			case <-done:
				fmt.Println("Exiting...")
				return
			}
		}
	}()

	log.Println(" [*] Waiting for logs. To exit press CTRL+C")
	wg.Wait()
}
