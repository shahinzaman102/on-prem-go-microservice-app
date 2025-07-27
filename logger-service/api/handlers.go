package api

import (
	"log-service/data"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	tracer := otel.Tracer("logger-service")
	_, span := tracer.Start(r.Context(), "WriteLog")
	defer span.End()

	var requestPayload JSONPayload
	_ = app.readJSON(w, r, &requestPayload)

	span.SetAttributes(
		attribute.String("log.name", requestPayload.Name),
		attribute.String("log.data", requestPayload.Data),
	)

	logger := Log.WithFields(logrus.Fields{
		"action": "WriteLog",
		"name":   requestPayload.Name,
		"data":   requestPayload.Data,
	})

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	duration := time.Since(start).Seconds()
	LogInsertionDuration.WithLabelValues("success").Observe(duration)

	if err != nil {
		logger.WithError(err).Error("Failed to insert log entry")
		LogInsertionErrors.Inc()
		LogInsertionDuration.WithLabelValues("failure").Observe(duration)

		span.RecordError(err)
		span.SetStatus(codes.Error, "MongoDB insert failed")

		app.errorJSON(w, err)
		return
	}

	logger.Info("Successfully logged entry")
	LogInsertionTotal.WithLabelValues("success").Inc()

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}
