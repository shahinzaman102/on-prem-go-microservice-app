package api

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectToMongo connects to the MongoDB instance and returns the client
func ConnectToMongo(mongoURL string) (*mongo.Client, error) {
	// Use the global logger from api.Log
	logger := Log.WithFields(logrus.Fields{
		"action":   "ConnectToMongo",
		"mongoURL": mongoURL,
	})

	// Create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// Try connecting to MongoDB
	logger.Info("Attempting to connect to MongoDB...")
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		// Log error if connection fails
		logger.WithError(err).Error("Error connecting to MongoDB")
		return nil, err
	}

	// Log success if connection is established
	logger.Info("Successfully connected to MongoDB")
	return c, nil
}
