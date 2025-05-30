package api

import (
	"fmt"
	"net/http"
	"net/smtp"

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

	// Mail API endpoint
	mux.Post("/send", app.SendMail)

	// Health Probes
	mux.Get("/liveness", app.LivenessProbe)
	mux.Get("/readiness", app.ReadinessProbe)

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

// LivenessProbe checks if the app is running
func (app *Config) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Liveness probe hit: /liveness")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessProbe checks if the SMTP service is reachable
func (app *Config) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	app.Logger.Info("Readiness probe hit: /readiness")

	// Type assertion: convert app.Mailer to *Mail
	mail, ok := app.Mailer.(*Mail)
	if !ok {
		app.Logger.Error("Readiness probe failed: unable to assert Mailer type")
		http.Error(w, "Readiness probe failed: unable to assert Mailer type", http.StatusInternalServerError)
		return
	}

	// Try connecting to the SMTP server
	err := testSMTPConnection(mail)
	if err != nil {
		app.Logger.WithError(err).Warn("Readiness probe failed: SMTP not reachable")
		http.Error(w, fmt.Sprintf("Readiness probe failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	app.Logger.Info("Readiness probe passed: SMTP is reachable")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// TestSMTPConnection tests if the SMTP server is reachable
func testSMTPConnection(mail *Mail) error {
	smtpAddress := fmt.Sprintf("%s:%d", mail.Host, mail.Port)

	// Attempt to connect to the SMTP server
	client, err := smtp.Dial(smtpAddress)
	if err != nil {
		return fmt.Errorf("error connecting to SMTP server: %v", err)
	}
	defer client.Quit()

	// Send a basic HELO command to check the server
	if err := client.Hello(mail.Host); err != nil {
		return fmt.Errorf("error sending HELO command: %v", err)
	}

	return nil
}
