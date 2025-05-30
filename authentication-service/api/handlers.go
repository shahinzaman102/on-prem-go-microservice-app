package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	logger := logrus.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		logger.WithError(err).Error("Failed to parse request payload")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		logger.WithError(err).Warn("Invalid credentials")
		app.Metrics.ErrorCount.WithLabelValues(r.Method, "/authenticate").Inc()
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		logger.Warn("Invalid password")
		app.Metrics.ErrorCount.WithLabelValues(r.Method, "/authenticate").Inc()
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// Log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		logger.WithError(err).Error("Failed to log authentication event")
		app.Metrics.ErrorCount.WithLabelValues(r.Method, "/authenticate").Inc()
		app.errorJSON(w, err)
		return
	}

	// Log successful authentication
	logger.WithFields(logrus.Fields{
		"user_email": user.Email,
	}).Info("User authenticated")

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)

	// Update Prometheus metrics
	duration := time.Since(start).Seconds()
	app.Metrics.RequestCount.WithLabelValues(r.Method, "/authenticate").Inc()
	app.Metrics.RequestLatency.WithLabelValues(r.Method, "/authenticate").Observe(duration)
	logger.WithField("latency", duration).Info("Request completed")
}

func (app *Config) logRequest(name, data string) error {
	logger := logrus.WithFields(logrus.Fields{
		"event": name,
		"data":  data,
	})

	logger.Info("Event logged")

	// Structure log data for external logging service
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		logger.WithError(err).Error("Failed to marshal log data to JSON")
		return err
	}

	logServiceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithError(err).Error("Failed to create HTTP request for logger service")
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logger.WithError(err).Error("Failed to send log data to logger service")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		err = fmt.Errorf("logger service returned status code %d", response.StatusCode)
		logger.WithError(err).Error("Logger service error")
		return err
	}

	return nil
}
