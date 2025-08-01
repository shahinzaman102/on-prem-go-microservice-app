package main

import (
	"authentication/api"
	"authentication/data"
	"authentication/internal/tracing"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const webPort = "80"
const maxDBRetries = 10 // Maximum retry attempts for DB connection

var counts int64
var logger = logrus.New()

// Initialize the app variable, but it will be updated later after database connection is successful
var app *api.Config

func main() {
	// Configure the logger
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting authentication service")

	// Connect to DB
	conn := connectToDB()
	if conn == nil {
		logger.Fatal("Unable to establish database connection after maximum retries")
	}

	// Initialize the app with the connection
	app = &api.Config{
		DB:     conn,
		Models: data.New(conn),
		Logger: logger,
		Redis:  initRedis(),
	}

	// Initialize Prometheus metrics
	app.Metrics.RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_request_count",
			Help: "Total number of requests received",
		},
		[]string{"method", "endpoint"},
	)

	app.Metrics.RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_service_request_latency_seconds",
			Help:    "Latency of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	app.Metrics.ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_error_count",
			Help: "Total number of errors",
		},
		[]string{"method", "endpoint"},
	)

	// Define PostgreSQL Connection Status Metric
	app.Metrics.PGConnectionStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_connection_status",
			Help: "up if PostgreSQL connection is up, down if down",
		},
		[]string{"job"},
	)

	// Register Prometheus metrics
	prometheus.MustRegister(app.Metrics.RequestCount)
	prometheus.MustRegister(app.Metrics.RequestLatency)
	prometheus.MustRegister(app.Metrics.ErrorCount)
	prometheus.MustRegister(app.Metrics.PGConnectionStatus)

	// Add /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	ctx := context.Background()

	// Initialize OpenTelemetry
	shutdown, err := tracing.InitTracer(ctx, "authentication-service")
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize tracer")
	}
	defer shutdown(ctx)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	logger.Info("Starting server on port ", webPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.WithError(err).Error("Failed to open DB connection")
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.WithError(err).Error("Failed to ping DB")
		return nil, err
	}

	return db, nil
}

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // assuming Redis service in Docker/K8s
		Password: "",           // no password
		DB:       0,
	})

	// Ping test
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	logger.Info("Connected to Redis")
	return rdb
}

func connectToDB() *sql.DB {
	// Read DSN from environment variables
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")

	// ⛔ safety check
	if postgresUser == "" || postgresPassword == "" || postgresDB == "" {
		logger.Fatalf("Missing required ENV vars: POSTGRES_USER=%q, POSTGRES_PASSWORD=%q, POSTGRES_DB=%q",
			postgresUser, postgresPassword, postgresDB)
	}

	// ✅ Log safe DSN (don't log password)
	logger.Infof("Using DB config: user=%s db=%s", postgresUser, postgresDB)

	// Construct DSN dynamically
	dsn := fmt.Sprintf("host=postgres port=5432 user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5",
		postgresUser, postgresPassword, postgresDB)

	// Log the DSN value (Be cautious with logging sensitive data)
	logger.Infof("Constructed DSN: %s", dsn)

	// Retry logic
	for {
		connection, err := openDB(dsn)
		if err != nil {
			logger.WithError(err).Warn("Postgres not ready...")
			counts++
			// Update PostgreSQL connection status to down (0)
			if app != nil {
				app.Metrics.PGConnectionStatus.WithLabelValues("postgres").Set(0)
			}
		} else {
			logger.Info("Connected to Postgres!")
			// Update PostgreSQL connection status to up (1)
			if app != nil {
				app.Metrics.PGConnectionStatus.WithLabelValues("postgres").Set(1)
			}
			return connection
		}

		if counts >= maxDBRetries {
			logger.WithFields(logrus.Fields{
				"max_retries": maxDBRetries,
			}).Error("Exceeded maximum DB connection attempts")
			return nil
		}

		// Retry logic with delay
		logger.WithFields(logrus.Fields{
			"retry":       counts + 1,
			"max_retries": maxDBRetries,
		}).Info("Retrying DB connection in 2 seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
