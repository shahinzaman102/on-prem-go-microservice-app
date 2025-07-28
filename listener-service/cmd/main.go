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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	prometheus.MustRegister(RabbitConnectionAttempts)
	prometheus.MustRegister(RabbitConnectionErrors)
}

func main() {
	// Start metrics server
	go startMetricsServer()

	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	ctx := context.Background()

	// Tracing setup
	shutdown, err := tracing.InitTracer(ctx, "listener-service")
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("listener-service")
	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	// Connect to RabbitMQ with tracing
	rabbitConn, err := connect(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to connect to RabbitMQ")
		os.Exit(1)
	}
	defer rabbitConn.Close()

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready(rabbitConn))

	log.Info("Listening for and consuming RabbitMQ messages...")

	// Trace consumer creation
	ctx, consumerSpan := tracer.Start(ctx, "create_consumer")
	consumer, err := event.NewConsumer(rabbitConn, log)
	consumerSpan.End()
	if err != nil {
		log.WithError(err).Fatal("Failed to create RabbitMQ consumer")
	}

	// Trace message consumption
	ctx, listenSpan := tracer.Start(ctx, "consume_events")
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	listenSpan.End()
	if err != nil {
		log.WithError(err).Error("Error while consuming messages")
	}
}

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	addr := ":80"
	log.Infof("Starting metrics server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting metrics server: %v", err)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Liveness probe passed: /healthz")
}

func ready(rabbitConn *amqp.Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rabbitConn.IsClosed() {
			http.Error(w, "RabbitMQ not connected", http.StatusServiceUnavailable)
			log.Warn("Readiness probe failed: RabbitMQ is not connected")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		log.Info("Readiness probe passed: /ready")
	}
}

// connect handles RabbitMQ connection with retries
func connect(ctx context.Context) (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	tracer := otel.Tracer("listener-service")
	ctx, span := tracer.Start(ctx, "connect_to_rabbitmq")
	defer span.End()

	for {
		RabbitConnectionAttempts.Inc()
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			RabbitConnectionErrors.Inc()
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
		span.SetAttributes(attribute.Int64("retry_attempts", counts))
		log.WithField("backOff", backOff).Info("Backing off before retry")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
