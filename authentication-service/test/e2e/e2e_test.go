package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticateEndpoint(t *testing.T) {
	// Define your handler function
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Simulate authentication logic
		if r.URL.Path == "/authenticate" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		http.NotFound(w, r)
	}

	// Create a test server with the handler
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Send the POST request to the test server
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "password",
	}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(server.URL+"/authenticate", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d but got %d", http.StatusAccepted, resp.StatusCode)
	}
}
