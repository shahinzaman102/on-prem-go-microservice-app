package api

import (
	"log-service/data"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// Config struct holds the models and the MongoDB client
type Config struct {
	Models data.Models
	Client *mongo.Client
	Logger *logrus.Logger
}
