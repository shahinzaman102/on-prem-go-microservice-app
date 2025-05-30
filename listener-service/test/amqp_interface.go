package event

import "github.com/rabbitmq/amqp091-go"

// AMQPChannelInterface abstracts RabbitMQ channel operations
type AMQPChannelInterface interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
}
