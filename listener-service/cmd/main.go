package main

import (
	"context"
	"listener/event"
	"listener/internal/tracing"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()

	RabbitConnectionAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rabbitmq_connection_attempts_total",
			Help: "Total number of RabbitMQ connection attempts.",
		},
	)

	RabbitConnectionErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rabbitmq_connection_errors_total",
			Help: "Total number of RabbitMQ connection errors.",
		},
	)
)

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(RabbitConnectionAttempts)
	prometheus.MustRegister(RabbitConnectionErrors)
}

func main() {
	// Start HTTP server for Prometheus metrics
	go startMetricsServer()

	// Configure logrus for JSON output
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// Initialize OpenTelemetry
	shutdown, err := tracing.InitTracer(context.Background(), "listener-service")
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer shutdown(context.Background())

	// Try to connect to RabbitMQ
	rabbitConn, err := connect()
	if err != nil {
		log.WithError(err).Error("Failed to connect to RabbitMQ")
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// Expose readiness probe
	http.HandleFunc("/healthz", healthz)         // Liveness Probe
	http.HandleFunc("/ready", ready(rabbitConn)) // Readiness Probe

	log.Info("Listening for and consuming RabbitMQ messages...")

	// Create consumer
	consumer, err := event.NewConsumer(rabbitConn, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create RabbitMQ consumer")
	}

	// Watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.WithError(err).Error("Error while consuming messages")
	}
}

// startMetricsServer launches an HTTP server to expose metrics
func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	addr := ":80" // Ensure this matches the port in Kubernetes config
	log.Infof("Starting metrics server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting metrics server: %v", err)
	}
}

// healthz is the Liveness Probe endpoint
func healthz(w http.ResponseWriter, r *http.Request) {
	// Here we return a 200 OK if the service is alive
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Liveness probe passed: /healthz")
}

// ready is the Readiness Probe endpoint
func ready(rabbitConn *amqp.Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if RabbitMQ is connected
		if rabbitConn.IsClosed() {
			// If RabbitMQ is not connected, return a 500 status indicating not ready
			http.Error(w, "RabbitMQ not connected", http.StatusServiceUnavailable)
			log.Warn("Readiness probe failed: RabbitMQ is not connected")
			return
		}

		// If RabbitMQ is connected, return 200 OK indicating readiness
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		log.Info("Readiness probe passed: /ready")
	}
}

// connect handles RabbitMQ connection with retries
func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		RabbitConnectionAttempts.Inc() // Increment connection attempt counter
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			RabbitConnectionErrors.Inc() // Increment connection error counter
			log.WithError(err).Warn("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Info("Connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			log.WithError(err).Error("Exceeded maximum retry attempts")
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.WithField("backOff", backOff).Info("Backing off before retry")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
