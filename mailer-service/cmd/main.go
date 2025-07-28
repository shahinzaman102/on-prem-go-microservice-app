package main

import (
	"context"
	"fmt"
	"mailer-service/api"
	"mailer-service/internal/tracing"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
)

const webPort = "80"

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	shutdown, err := tracing.InitTracer(ctx, "mailer-service") // ✅ FIXED service name
	if err != nil {
		fmt.Println("Failed to initialize tracer:", err)
		os.Exit(1)
	}
	defer shutdown(ctx)

	tracer := otel.Tracer("mailer-service")
	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	app := api.Config{}

	// Initialize logger
	app.InitializeLogger()

	// Trace mail config setup
	ctx, mailSpan := tracer.Start(ctx, "create_mail_config")
	app.Mailer = createMail()
	mailSpan.End()

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		app.Logger.Info("Starting Prometheus metrics server on :9913")
		if err := http.ListenAndServe(":9913", nil); err != nil {
			app.Logger.Fatal("Error starting Prometheus metrics server:", err)
		}
	}()

	// Trace web server startup
	_, srvSpan := tracer.Start(ctx, "start_http_server")
	app.Logger.Info("Starting mail service on port " + webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}
	srvSpan.End()

	err = srv.ListenAndServe()
	if err != nil {
		app.Logger.Panic("Error starting mail service:", err)
	}
}

// Create Mail setup for SMTP
func createMail() *api.Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := &api.Mail{
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
