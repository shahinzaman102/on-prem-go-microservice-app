package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// Explicitly define health check routes
	mux.Get("/healthz", app.HealthzHandler)
	mux.Get("/ready", app.ReadinessHandler)

	// Application-specific routes
	mux.Post("/log", app.WriteLog)

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

	// Check if MongoDB is connected
	if app.Client.Ping(r.Context(), nil) != nil {
		app.Logger.Warn("Readiness probe failed: MongoDB is not connected")
		http.Error(w, "MongoDB not connected", http.StatusServiceUnavailable)
		return
	}

	app.Logger.Info("Readiness probe passed: /ready")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
