package api

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// InitLogger initializes the logger
func InitLogger() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{}) // Ensure JSON format
	Log.SetOutput(os.Stdout)                  // Ensure logs are written to stdout
	Log.SetLevel(logrus.InfoLevel)            // Set the log level (info or higher)
}
