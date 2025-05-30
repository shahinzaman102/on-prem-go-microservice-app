package unit

// import (
// 	"context"
// 	"log-service/api"
// 	"log-service/data"
// 	"log-service/logs"
// 	"log-service/test/unit/mock_data"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestWriteLogGRPC(t *testing.T) {
// 	// Create a new mock for LogEntry
// 	mockLogEntry := new(mock_data.MockLogEntry)
// 	mockLogEntry.On("Insert", mock.Anything).Return(nil).Once()

// 	// Set up the LogServer with the mock LogEntry
// 	server := &api.LogServer{
// 		Models: data.Models{
// 			LogEntry: mockLogEntry,
// 		},
// 	}

// 	req := &logs.LogRequest{
// 		LogEntry: &logs.Log{
// 			Name: "Test Name",
// 			Data: "Test Data",
// 		},
// 	}

// 	// Call WriteLog method via gRPC
// 	resp, err := server.WriteLog(context.Background(), req)

// 	// Assert results
// 	mockLogEntry.AssertExpectations(t)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "logged!", resp.GetResult())
// }
