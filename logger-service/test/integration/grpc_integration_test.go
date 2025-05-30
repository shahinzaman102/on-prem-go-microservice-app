package integration

import (
	"fmt"
	"testing"
	"time"
)

func TestGRPCServer(t *testing.T) {
	listener, err := getListener()
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	server := newServer()

	// Create a channel to communicate errors from the goroutine to the main goroutine
	errChan := make(chan error, 1)

	go func() {
		if err := server.Serve(listener); err != nil {
			// Send the error to the main goroutine
			errChan <- fmt.Errorf("Failed to serve: %v", err)
		}
	}()

	// Check for any errors from the goroutine
	select {
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(time.Second): // Optional: Add a timeout
		// Continue with the rest of your test
	}

	// Optionally, add some logic to close the server after the test is complete
	defer server.Stop()
}
