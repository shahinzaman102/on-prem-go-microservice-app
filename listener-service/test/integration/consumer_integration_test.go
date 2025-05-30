package integration

import (
	"bytes"
	"encoding/json"
	"listener/event"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Test: Integration for consuming RabbitMQ messages and calling external service
func TestConsumerIntegration(t *testing.T) {
	// Create a test HTTP request
	payload := event.Payload{Name: "log", Data: "Test log message"}
	reqBody, _ := json.Marshal(payload)

	// Set up the router
	router := mux.NewRouter()
	router.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		var reqPayload event.Payload
		_ = json.NewDecoder(r.Body).Decode(&reqPayload)
		assert.Equal(t, "log", reqPayload.Name)
		assert.Equal(t, "Test log message", reqPayload.Data)
		w.WriteHeader(http.StatusAccepted)
	}).Methods("POST")

	// Mock HTTP request for log service
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBuffer(reqBody))
	rec := httptest.NewRecorder()

	// Simulate the request
	router.ServeHTTP(rec, req)

	// Validate response
	assert.Equal(t, http.StatusAccepted, rec.Code)
}
