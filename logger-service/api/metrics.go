package api

import "github.com/prometheus/client_golang/prometheus"

var (
	// Track the number of log insertions
	LogInsertionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logger_service_log_inserts_total",
			Help: "Total number of log entries inserted into the database.",
		},
		[]string{"status"}, // You can add "status" label to track success/failure
	)

	// Track the number of log insertion errors
	LogInsertionErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "logger_service_log_insertion_errors_total",
			Help: "Total number of errors occurred while inserting logs.",
		},
	)

	// Track the duration of the log insertion operations
	LogInsertionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "logger_service_log_insertion_duration_seconds",
			Help:    "Histogram of log insertion durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"}, // You can add status label to track success/failure
	)

	// Track gRPC request durations
	GrpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "logger_service_grpc_request_duration_seconds",
			Help:    "Histogram of gRPC request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"}, // Track method for each gRPC call
	)
)

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(LogInsertionTotal)
	prometheus.MustRegister(LogInsertionErrors)
	prometheus.MustRegister(LogInsertionDuration)
	prometheus.MustRegister(GrpcRequestDuration)
}
