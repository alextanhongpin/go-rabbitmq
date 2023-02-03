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

func main() {
	// 1. Dial rabbitmq.
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

	// 3. Delcare a queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		QueueName: "hello",
	})
	if err != nil {
		log.Panicf("failed to declare queue: %v", err)
	}

	// 4. Create a consumer for the queue.
	msgs, err := amqplib.Consume(ch, amqplib.ConsumeOption{
		QueueName: q.Name,
		AutoAck:   true,
	})
	if err != nil {
		log.Panicf("failed to register consumer: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	// 5. Read messages.
	go func() {
		defer wg.Done()

		for {
			select {
			case msg := <-msgs:
				fmt.Println("Received a message:", string(msg.Body))
			case <-done:
				return
			}
		}
	}()

	log.Println("[*] Waiting for messages. Press ctrl + c to exit.")
	wg.Wait()
}
