package e2e

// import (
// 	"bytes"
// 	"encoding/json"
// 	"log-service/api"
// 	"log-service/data"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// // MockLogEntry is a mock implementation of the LogEntryInterface
// type MockLogEntry struct{}

// // Mock Insert method
// func (m *MockLogEntry) Insert(entry data.LogEntry) error {
// 	// Simulate successful insertion
// 	return nil
// }

// // Mock All method
// func (m *MockLogEntry) All() ([]*data.LogEntry, error) {
// 	// Return a mock list of log entries
// 	return []*data.LogEntry{
// 		{Name: "Mock Name", Data: "Mock Data"},
// 	}, nil
// }

// // Mock GetOne method
// func (m *MockLogEntry) GetOne(id string) (*data.LogEntry, error) {
// 	// Return a single mock log entry
// 	return &data.LogEntry{Name: "Mock Name", Data: "Mock Data"}, nil
// }

// // Mock DropCollection method
// func (m *MockLogEntry) DropCollection() error {
// 	// Simulate successful collection drop
// 	return nil
// }

// // Mock Update method
// func (m *MockLogEntry) Update() (*mongo.UpdateResult, error) {
// 	// Simulate a successful update
// 	return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
// }

// // TestCreateLogEntry tests the WriteLog handler
// func TestCreateLogEntry(t *testing.T) {
// 	// Setup mock dependencies
// 	mockLogEntry := &MockLogEntry{}
// 	mockModels := data.Models{
// 		LogEntry: mockLogEntry, // Inject the mock
// 	}

// 	// Initialize the Config object with the mock Models
// 	app := api.Config{
// 		Models: mockModels,
// 	}

// 	// Prepare mock payload
// 	payload := api.JSONPayload{
// 		Name: "E2E Test Name",
// 		Data: "E2E Test Data",
// 	}

// 	payloadBytes, _ := json.Marshal(payload)
// 	req := httptest.NewRequest("POST", "/log", bytes.NewReader(payloadBytes))
// 	rec := httptest.NewRecorder()

// 	// Call the WriteLog handler
// 	app.WriteLog(rec, req)

// 	// Assert response
// 	assert.Equal(t, http.StatusAccepted, rec.Code)
// 	assert.Contains(t, rec.Body.String(), "logged")
// }
