package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func TestMongoInsert(t *testing.T) {
	// Create a new mock MongoDB collection
	mockCollection := new(MockCollection)

	// Define the test data
	data := map[string]interface{}{
		"name": "Integration Test Name",
		"data": "Integration Test Data",
	}

	// Set up expectations for the InsertOne call
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil)

	// Call the function you want to test, which uses the mock collection
	result, err := mockCollection.InsertOne(context.TODO(), data)

	// Assert no error occurred
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Verify that the InsertOne method was called
	mockCollection.AssertExpectations(t)
}
