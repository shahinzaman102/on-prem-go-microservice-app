package unit

// import (
// 	"listener/event"
// 	"testing"

// 	"github.com/rabbitmq/amqp091-go"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// // MockAMQPChannel implements event.AMQPChannelInterface
// type MockAMQPChannel struct {
// 	mock.Mock
// }

// func (m *MockAMQPChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
// 	argsList := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
// 	return argsList.Error(0)
// }

// func (m *MockAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
// 	argsList := m.Called(name, durable, autoDelete, exclusive, noWait, args)
// 	return argsList.Get(0).(amqp091.Queue), argsList.Error(1)
// }

// func TestDeclareExchange(t *testing.T) {
// 	mockChannel := new(MockAMQPChannel)

// 	mockChannel.On("ExchangeDeclare", "logs_topic", "topic", true, false, false, false, mock.AnythingOfType("amqp091.Table")).Return(nil)

// 	err := event.DeclareExchange(mockChannel)
// 	assert.NoError(t, err)
// 	mockChannel.AssertExpectations(t)
// }

// func TestDeclareRandomQueue(t *testing.T) {
// 	mockChannel := new(MockAMQPChannel)

// 	mockQueue := amqp091.Queue{Name: "test-queue"}
// 	mockChannel.On("QueueDeclare", "", false, false, true, false, mock.AnythingOfType("amqp091.Table")).Return(mockQueue, nil)

// 	queue, err := event.DeclareRandomQueue(mockChannel)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test-queue", queue.Name)
// 	mockChannel.AssertExpectations(t)
// }
