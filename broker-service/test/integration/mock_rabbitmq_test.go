package integration

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

// MockRabbitMQConnection is a mock implementation of the RabbitMQ connection.
type MockRabbitMQConnection struct {
	mock.Mock
}

// Channel is a mocked method to simulate getting a RabbitMQ channel.
func (m *MockRabbitMQConnection) Channel() (*amqp091.Channel, error) {
	args := m.Called()
	return args.Get(0).(*amqp091.Channel), args.Error(1)
}

// Publish is a mocked method to simulate publishing a message.
func (m *MockRabbitMQConnection) Publish(queue string, body []byte) error {
	args := m.Called(queue, body)
	return args.Error(0)
}

// Close is a mocked method to simulate closing the connection.
func (m *MockRabbitMQConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}
