package e2e

// import (
// 	"encoding/json"
// 	"listener/event"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestListenerE2E(t *testing.T) {
// 	// Mock the payload
// 	payload := event.Payload{
// 		Name: "log",
// 		Data: "E2E test message",
// 	}

// 	// Mock the external log service
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var reqPayload event.Payload
// 		err := json.NewDecoder(r.Body).Decode(&reqPayload)
// 		assert.NoError(t, err)

// 		// Verify the payload
// 		assert.Equal(t, "log", reqPayload.Name)
// 		assert.Equal(t, "E2E test message", reqPayload.Data)

// 		w.WriteHeader(http.StatusAccepted)
// 	}))
// 	defer server.Close()

// 	// Override the log service URL
// 	originalLogServiceURL := event.LogServiceURL
// 	event.LogServiceURL = server.URL
// 	defer func() { event.LogServiceURL = originalLogServiceURL }()

// 	// Call the LogEvent function
// 	err := event.LogEvent(payload)
// 	assert.NoError(t, err)
// }
