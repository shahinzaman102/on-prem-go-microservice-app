package main

import (
	"broker/api"
	"broker/internal/tracing"
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const webPort = "8080"

// Initialize logger
var logger = logrus.New()

func main() {
	// Configure logrus for JSON output
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// try to connect to rabbitmq
	rabbitConn, err := connect(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to RabbitMQ")
	}
	defer rabbitConn.Close()

	app := api.Config{
		Rabbit: rabbitConn,
		Logger: logger,
	}

	// Start logging the application initialization
	logger.Infof("Starting broker service on port %s", webPort)

	shutdown, err := tracing.InitTracer(context.Background(), "broker-service")
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer shutdown(context.Background()) // clean shutdown on exit

	// Define HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	// Start the server and log if there's an error
	err = srv.ListenAndServe()
	if err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}
}

// Connect to RabbitMQ with retry logic
func connect(logger *logrus.Logger) (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// Don't continue until RabbitMQ is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			logger.Warn("RabbitMQ not yet ready...")
			counts++
		} else {
			logger.Info("Connected to RabbitMQ!")
			connection = c
			break
		}

		// If we've retried more than 5 times, exit
		if counts > 5 {
			logger.WithError(err).Error("Failed to connect to RabbitMQ after several attempts")
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		logger.WithFields(logrus.Fields{
			"retry_count":  counts,
			"backoff_time": backOff,
		}).Info("Backing off before retrying RabbitMQ connection")

		// Sleep before retrying
		time.Sleep(backOff)
	}

	return connection, nil
}
