package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alextanhongpin/go-rabbitmq/amqplib"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var queue string
	flag.StringVar(&queue, "queue", "hello", "queue name")
	flag.Parse()

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

	// 3. Declare a durable queue.
	q, err := amqplib.QueueDeclare(ch, amqplib.QueueDeclareOption{
		QueueName: queue,
		Durable:   true,
	})
	if err != nil {
		log.Panicf("failed to declare queue: %v", err)
	}

	// 4. Set the QoS.
	err = amqplib.Qos(ch, amqplib.QosOption{
		PrefetchSize: 1,
	})
	if err != nil {
		log.Panicf("failed to set qos: %v", err)
	}

	// 5. Create a consumer for the queue with manual acknowledgement.
	msgs, err := amqplib.Consume(ch, amqplib.ConsumeOption{
		QueueName: q.Name,
		AutoAck:   false,
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case msg := <-msgs:
				fmt.Println("Received a message:", string(msg.Body))
				t := bytes.Count(msg.Body, []byte("."))
				time.Sleep(time.Duration(t) * time.Second)
				fmt.Println("Done")

				multiple := false
				msg.Ack(multiple)
			case <-done:
				return
			}
		}
	}()

	log.Println("[*] Waiting for messages. Press ctrl + c to exit.")
	wg.Wait()
}
