package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (app *Config) Routes() http.Handler {
	mux := chi.NewRouter()

	// Traced routes
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(app.Broker), "Broker"))
	mux.Handle("/handle", otelhttp.NewHandler(http.HandlerFunc(app.HandleSubmission), "HandleSubmission"))
	mux.Handle("/log-grpc", otelhttp.NewHandler(http.HandlerFunc(app.LogViaGRPC), "LogViaGRPC"))

	// specify who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check routes
	mux.Get("/healthz", app.HealthzHandler)
	mux.Get("/ready", app.ReadinessHandler)

	return mux
}

// HealthzHandler for liveness probe
func (app *Config) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Liveness probe hit: /healthz")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessHandler for readiness probe
func (app *Config) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Readiness probe hit: /ready")

	if app.Rabbit == nil || app.Rabbit.IsClosed() {
		app.Logger.Warn("Readiness probe failed: RabbitMQ is not connected")
		http.Error(w, "RabbitMQ not connected", http.StatusServiceUnavailable)
		return
	}

	app.Logger.Info("Readiness probe passed: /ready")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
