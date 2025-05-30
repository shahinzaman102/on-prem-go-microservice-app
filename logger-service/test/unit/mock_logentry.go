package unit

// import (
// 	"log-service/data"

// 	"github.com/stretchr/testify/mock"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// // MockLogEntry is a mock implementation of the LogEntryInterface
// type MockLogEntry struct {
// 	mock.Mock
// }

// // Insert is a mocked method for inserting a log entry
// func (m *MockLogEntry) Insert(entry data.LogEntry) error {
// 	args := m.Called(entry)
// 	return args.Error(0)
// }

// // All is a mocked method to return all log entries
// func (m *MockLogEntry) All() ([]*data.LogEntry, error) {
// 	args := m.Called()
// 	return args.Get(0).([]*data.LogEntry), args.Error(1)
// }

// // GetOne is a mocked method to get one log entry by ID
// func (m *MockLogEntry) GetOne(id string) (*data.LogEntry, error) {
// 	args := m.Called(id)
// 	return args.Get(0).(*data.LogEntry), args.Error(1)
// }

// // DropCollection is a mocked method to drop the log collection
// func (m *MockLogEntry) DropCollection() error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// // Update is a mocked method to update a log entry
// func (m *MockLogEntry) Update() (*mongo.UpdateResult, error) {
// 	args := m.Called()
// 	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
// }
