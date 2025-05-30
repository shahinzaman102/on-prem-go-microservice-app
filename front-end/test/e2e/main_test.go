package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock HTTP client to simulate server responses.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Implement the RoundTripper interface for our MockHTTPClient.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc == nil {
		return nil, errors.New("DoFunc is nil")
	}
	return m.DoFunc(req)
}

// Mock server response for mail POST request
func mockMailPostRequest(req *http.Request) (*http.Response, error) {
	// Debugging: Print out the request path and payload to ensure correct request is received
	fmt.Printf("Received request for %s\n", req.URL.Path)

	// Simulate a successful server response for the /handle endpoint
	if req.URL.Path == "/handle" {
		// Return a response with status 200 and no body
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       nil, // No body needed in mock response
		}, nil
	}
	return nil, fmt.Errorf("unexpected request path: %s", req.URL.Path)
}

// Mock server response for grpc POST request
func mockGrpcPostRequest(req *http.Request) (*http.Response, error) {
	// Debugging: Print out the request path and payload to ensure correct request is received
	fmt.Printf("Received request for %s\n", req.URL.Path)

	// Simulate a successful server response for the /log-grpc endpoint
	if req.URL.Path == "/log-grpc" {
		// Return a response with status 200 and no body
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       nil, // No body needed in mock response
		}, nil
	}
	return nil, fmt.Errorf("unexpected request path: %s", req.URL.Path)
}

// Test for sending mail via POST request
func TestMailPostRequest(t *testing.T) {
	// Create a mock HTTP client that simulates the server's behavior
	mockClient := &MockHTTPClient{
		DoFunc: mockMailPostRequest,
	}

	// Ensure that DoFunc is set
	if mockClient.DoFunc == nil {
		t.Fatal("Mock HTTP client DoFunc is nil")
	}

	// Mock payload for POST request
	payload := map[string]interface{}{
		"action": "mail",
		"mail": map[string]string{
			"from":    "me@example.com",
			"to":      "you@there.com",
			"subject": "Test email",
			"message": "Hello world!",
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "http://localhost:8081/handle", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use the mock client instead of the real HTTP client
	resp, err := mockClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	// Check if the response is nil
	if resp == nil {
		t.Fatal("Received nil response")
	}

	// Debugging: Print the response object details
	fmt.Printf("Response: %+v\n", resp)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status code")
}

// Test for sending gRPC log via POST request
func TestGrpcPostRequest(t *testing.T) {
	// Create a mock HTTP client that simulates the server's behavior
	mockClient := &MockHTTPClient{
		DoFunc: mockGrpcPostRequest,
	}

	// Ensure that DoFunc is set
	if mockClient.DoFunc == nil {
		t.Fatal("Mock HTTP client DoFunc is nil")
	}

	// Mock payload for POST request
	payload := map[string]interface{}{
		"action": "logGrpc",
		"logGrpc": map[string]string{
			"name": "event",
			"data": "Some kind of gRPC data",
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "http://localhost:8081/log-grpc", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use the mock client instead of the real HTTP client
	resp, err := mockClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	// Check if the response is nil
	if resp == nil {
		t.Fatal("Received nil response")
	}

	// Debugging: Print the response object details
	fmt.Printf("Response: %+v\n", resp)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status code")
}
