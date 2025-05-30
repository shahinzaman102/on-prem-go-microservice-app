package unit

// import (
// 	"log-service/api"
// 	"log-service/data"
// 	"log-service/test/unit/mock_data"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestWriteLog(t *testing.T) {
// 	// Create a new mock for LogEntry
// 	mockLogEntry := new(mock_data.MockLogEntry)
// 	mockLogEntry.On("Insert", mock.Anything).Return(nil).Once()

// 	// Set up the config with the mock LogEntry
// 	app := api.Config{
// 		Models: data.Models{
// 			LogEntry: mockLogEntry, // Pass the mock here
// 		},
// 	}

// 	// Create a new request with JSON payload
// 	req := httptest.NewRequest(http.MethodPost, "/log", nil)
// 	req.Header.Set("Content-Type", "application/json")

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Call the handler
// 	handler := http.HandlerFunc(app.WriteLog)
// 	handler.ServeHTTP(rr, req)

// 	// Assert that the insert was called and no error occurred
// 	mockLogEntry.AssertExpectations(t)
// 	assert.Equal(t, http.StatusAccepted, rr.Code)
// }
