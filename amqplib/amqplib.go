package amqplib

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueDeclareOption struct {
	QueueName        string // The name of the queue.
	Durable          bool   // The queue is persisted on broker restart.
	DeleteWhenUnused bool
	Exclusive        bool // If true, then only one connection can be made to this queue.
	NoWait           bool
	Args             amqp.Table
}

func QueueDeclare(ch *amqp.Channel, o QueueDeclareOption) (amqp.Queue, error) {
	return ch.QueueDeclare(
		o.QueueName,
		o.Durable,
		o.DeleteWhenUnused,
		o.Exclusive,
		o.NoWait,
		o.Args,
	)
}

type PublishWithContextOption struct {
	ExchangeName string
	RoutingKey   string
	Mandatory    bool
	Immediate    bool
	Message      amqp.Publishing
}

func PublishWithContext(ctx context.Context, ch *amqp.Channel, o PublishWithContextOption) error {
	return ch.PublishWithContext(ctx,
		o.ExchangeName,
		o.RoutingKey,
		o.Mandatory,
		o.Immediate,
		o.Message,
	)
}

type ConsumeOption struct {
	QueueName string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

func Consume(ch *amqp.Channel, o ConsumeOption) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		o.QueueName,
		o.Consumer,
		o.AutoAck,
		o.Exclusive,
		o.NoLocal,
		o.NoWait,
		o.Args,
	)
}

type QosOption struct {
	// PrefetchCount of 1 tells RabbitMQ not to give more than 1 message to a
	// worker at a time.
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

// Qos is defined at the consumer side.
func Qos(ch *amqp.Channel, o QosOption) error {
	return ch.Qos(
		o.PrefetchCount,
		o.PrefetchSize,
		o.Global,
	)
}

type ExchangeType string

const (
	ExchangeTypeDirect  ExchangeType = "direct"
	ExchangeTypeTopic   ExchangeType = "topic"
	ExchangeTypeHeaders ExchangeType = "headers"
	ExchangeTypeFanout  ExchangeType = "fanout"
)

type ExchangeDeclareOption struct {
	ExchangeName string
	Type         ExchangeType
	Durable      bool
	AutoDeleted  bool
	Internal     bool
	NoWait       bool
	Args         amqp.Table
}

func ExchangeDeclare(ch *amqp.Channel, o ExchangeDeclareOption) error {
	return ch.ExchangeDeclare(
		o.ExchangeName,
		string(o.Type),
		o.Durable,
		o.AutoDeleted,
		o.Internal,
		o.NoWait,
		o.Args,
	)
}

type QueueBindOption struct {
	QueueName    string
	BindingKey   string
	ExchangeName string
	NoWait       bool
	Args         amqp.Table
}

func QueueBind(ch *amqp.Channel, o QueueBindOption) error {
	return ch.QueueBind(
		o.QueueName,
		o.BindingKey,
		o.ExchangeName,
		o.NoWait,
		o.Args,
	)
}
