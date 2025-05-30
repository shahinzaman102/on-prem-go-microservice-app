package api

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Mailer Mailer
	Logger *logrus.Logger // Add Logger field to Config struct
}

// Initialize logrus logger in the config setup
func (app *Config) InitializeLogger() {
	app.Logger = logrus.New()                        // Initialize logrus logger
	app.Logger.SetFormatter(&logrus.JSONFormatter{}) // Set JSON format for structured logs
}
