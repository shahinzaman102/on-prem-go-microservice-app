package integration

import (
	"bytes"
	"encoding/json"
	"mailer-service/api"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMail is a mock of the Mail type
type MockMail struct {
	// We use the interface instead of embedding `Mail`
	Mailer api.Mailer
}

func (m *MockMail) SendSMTPMessage(msg api.Message) error {
	// Simulating successful email send (no actual SMTP send)
	return nil
}

func (m *MockMail) BuildHTMLMessage(msg api.Message) (string, error) {
	// Return a simple HTML message string for testing, bypassing template rendering
	return "<html><body><p>Test email body</p></body></html>", nil
}

func (m *MockMail) BuildPlainTextMessage(msg api.Message) (string, error) {
	// Return a simple plain text message string for testing, bypassing template rendering
	return "Test email body", nil
}

func TestSendMailHandler(t *testing.T) {
	// Create a mock mailer and initialize the embedded api.Mail struct
	mockMailer := &MockMail{
		Mailer: &api.Mail{ // Explicitly mock Mail interface
			Host: "smtp.example.com",
			Port: 587,
		},
	}

	// Set up the app with the mock mailer
	app := &api.Config{
		Mailer: mockMailer, // Use the mock mailer
	}

	// Create a test server using the routes
	ts := httptest.NewServer(app.Routes())
	defer ts.Close()

	// Prepare the request payload
	payload := map[string]string{
		"from":    "sender@example.com",
		"to":      "receiver@example.com",
		"subject": "Test Email",
		"message": "This is a test message",
	}
	jsonPayload, _ := json.Marshal(payload)

	// Make the POST request to /send
	req, _ := http.NewRequest("POST", ts.URL+"/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	// Assert that there is no error and check the status code
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}
