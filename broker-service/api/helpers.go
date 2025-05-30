package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

func init() {
	// Set log format to JSON for structured logging
	log.SetFormatter(&logrus.JSONFormatter{})
	// Set log level
	log.SetLevel(logrus.InfoLevel)
}

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// readJSON tries to read the body of a request and converts it into JSON
func (app *Config) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte
	log.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL.Path,
	}).Info("Attempting to read JSON request body")

	// Limit the size of the request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method": r.Method,
			"url":    r.URL.Path,
			"error":  err.Error(),
		}).Error("Failed to decode JSON request body")
		return err
	}

	// Ensure there's no extra content in the request body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		log.WithFields(logrus.Fields{
			"method": r.Method,
			"url":    r.URL.Path,
			"error":  "body must have only a single JSON value",
		}).Error("Invalid JSON request body")
		return errors.New("body must have only a single JSON value")
	}

	log.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL.Path,
	}).Info("Successfully read JSON request body")

	return nil
}

// writeJSON takes a response status code and arbitrary data and writes a JSON response to the client
func (app *Config) WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	log.WithFields(logrus.Fields{
		"status": status,
	}).Info("Attempting to write JSON response")

	out, err := json.Marshal(data)
	if err != nil {
		log.WithFields(logrus.Fields{
			"status": status,
			"error":  err.Error(),
		}).Error("Failed to marshal response data to JSON")
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		log.WithFields(logrus.Fields{
			"status": status,
			"error":  err.Error(),
		}).Error("Failed to write JSON response")
		return err
	}

	log.WithFields(logrus.Fields{
		"status": status,
	}).Info("Successfully wrote JSON response")
	return nil
}

// errorJSON takes an error, and optionally a response status code, and generates and sends
// a JSON error response
func (app *Config) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	log.WithFields(logrus.Fields{
		"status": statusCode,
		"error":  err.Error(),
	}).Error("Sending error JSON response")

	var payload JsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.WriteJSON(w, statusCode, payload)
}
