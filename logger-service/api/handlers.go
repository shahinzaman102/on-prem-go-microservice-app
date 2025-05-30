package api

import (
	"log-service/data"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// Start timer to measure the duration of the HTTP request
	start := time.Now()

	// Read json into var
	var requestPayload JSONPayload
	_ = app.readJSON(w, r, &requestPayload)

	// Log the incoming request
	logger := Log.WithFields(logrus.Fields{
		"action": "WriteLog",
		"name":   requestPayload.Name,
		"data":   requestPayload.Data,
	})

	// Insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)

	// Record the duration of the HTTP request
	duration := time.Since(start).Seconds()
	LogInsertionDuration.WithLabelValues("success").Observe(duration)

	if err != nil {
		// Log failure and increment error counter for log insertion
		logger.WithError(err).Error("Failed to insert log entry")

		// Increment error counter for log insertion
		LogInsertionErrors.Inc()

		// Record the failure duration
		LogInsertionDuration.WithLabelValues("failure").Observe(duration)

		app.errorJSON(w, err)
		return
	}

	// Log success and increment success counter for log insertion
	logger.Info("Successfully logged entry")

	// Increment success counter for log insertion
	LogInsertionTotal.WithLabelValues("success").Inc()

	// Respond with success message
	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}
