package unit

import (
	"errors"
	"mailer-service/api"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMailer is a mock of the api.Mailer interface
type MockMailer struct {
	mock.Mock
}

// SendSMTPMessage mocks the SendSMTPMessage method of api.Mailer
func (m *MockMailer) SendSMTPMessage(msg api.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestSendMail_Success(t *testing.T) {
	// Create a new MockMailer and set it in the app configuration
	mockMailer := new(MockMailer)
	app := &api.Config{
		Mailer: mockMailer, // Now this works because Mailer is the interface
	}

	// Define the expected behavior of the mock
	mockMailer.On("SendSMTPMessage", mock.Anything).Return(nil)

	// Simulate a successful email sending
	msg := api.Message{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test",
		Data:    "Test Message",
	}

	// Call the SendSMTPMessage method
	err := app.Mailer.SendSMTPMessage(msg)

	// Assert that there is no error
	assert.NoError(t, err)

	// Ensure that the mock was called as expected
	mockMailer.AssertExpectations(t)
}

func TestSendMail_Failure(t *testing.T) {
	// Create a new MockMailer and set it in the app configuration
	mockMailer := new(MockMailer)
	app := &api.Config{
		Mailer: mockMailer, // Now this works because Mailer is the interface
	}

	// Define the expected behavior of the mock to return an error
	mockMailer.On("SendSMTPMessage", mock.Anything).Return(errors.New("not implemented"))

	// Simulate a failed email sending
	msg := api.Message{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test",
		Data:    "Test Message",
	}

	// Call the SendSMTPMessage method
	err := app.Mailer.SendSMTPMessage(msg)

	// Assert that there is an error
	assert.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())

	// Ensure that the mock was called as expected
	mockMailer.AssertExpectations(t)
}
