package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *Config) Routes() http.Handler {
	mux := chi.NewRouter()

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
	mux.Use(app.MetricsMiddleware)

	// Add metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoints
	mux.Get("/healthz", app.HealthzHandler)     // Liveness probe
	mux.Get("/readiness", app.ReadinessHandler) // Readiness probe

	mux.Post("/authenticate", app.Authenticate)
	return mux
}

// HealthzHandler for liveness probe
func (app *Config) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Healthz endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessHandler for readiness probe
func (app *Config) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Readiness endpoint hit")

	if app.DB == nil {
		app.Logger.Warn("DB connection is not available")
		http.Error(w, "Service not ready, DB connection is down", http.StatusServiceUnavailable)
		return
	}

	// Try to ping the database to ensure readiness
	err := app.DB.Ping()
	if err != nil {
		app.Logger.WithError(err).Warn("DB ping failed")
		http.Error(w, "Service not ready, DB ping failed", http.StatusServiceUnavailable)
		return
	}

	app.Logger.Info("DB is ready, responding OK")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Middleware to track request count and latency
func (app *Config) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment request count
		app.Metrics.RequestCount.WithLabelValues(r.Method, r.URL.Path).Inc()

		// Capture response status
		statusRecorder := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(statusRecorder, r)

		// Measure latency
		duration := time.Since(start).Seconds()
		app.Metrics.RequestLatency.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// Custom response writer to track status codes
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
