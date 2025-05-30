package integration

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestMockRabbitMQConnection(t *testing.T) {
	// Create a mock RabbitMQ connection
	mockConn := new(MockRabbitMQConnection)

	// Mock the Channel method to return a mock channel
	mockChannel := new(amqp091.Channel) // You can mock methods of this channel if needed
	mockConn.On("Channel").Return(mockChannel, nil)

	// Call the method under test (for example, a function that uses the connection)
	ch, err := mockConn.Channel()

	// Assert that the method behaves as expected
	assert.NoError(t, err)
	assert.Equal(t, mockChannel, ch)

	// Verify the mock expectations
	mockConn.AssertExpectations(t)
}
