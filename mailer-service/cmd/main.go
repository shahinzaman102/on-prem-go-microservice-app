package main

import (
	"context"
	"fmt"
	"mailer-service/api"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const webPort = "80"

func main() {
	app := api.Config{
		Mailer: createMail(), // Ensure it returns *api.Mail, not api.Mail
	}

	// Initialize logger
	app.InitializeLogger()

	// Initialize OpenTelemetry Tracer
	shutdown, err := api.InitTracer("mailer-service")
	if err != nil {
		app.Logger.Fatal("Failed to init tracer: ", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			app.Logger.Error("Error shutting down tracer: ", err)
		}
	}()

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		app.Logger.Info("Starting Prometheus metrics server on :9913") // Change port to 9913
		if err := http.ListenAndServe(":9913", nil); err != nil {
			app.Logger.Fatal("Error starting Prometheus metrics server:", err)
		}
	}()

	app.Logger.Info("Starting mail service on port " + webPort) // Use app.Logger for info logs

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		app.Logger.Panic("Error starting mail service:", err)
	}
}

// Create Mail setup for SMTP
func createMail() *api.Mail { // Return *api.Mail here
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := &api.Mail{ // Use a pointer to api.Mail
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
	}

	return m
}
