package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// readJSON tries to read the body of a request and converts it into JSON
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	// Set the maximum allowed size of the request body to 1MB
	maxBytes := 1048576 // one megabyte

	// Log request start for reading JSON
	logger := Log.WithFields(logrus.Fields{
		"action": "readJSON",
		"method": r.Method,
		"uri":    r.RequestURI,
	})

	// Limit the size of the request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Decode the JSON into the provided data object
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		logger.WithError(err).Error("Failed to decode JSON")
		return err
	}

	// Check for any extra data after the initial JSON object
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		logger.Error("Body must have only a single JSON value")
		return errors.New("body must have only a single JSON value")
	}

	logger.Info("Successfully read JSON")
	return nil
}

// writeJSON takes a response status code and arbitrary data and writes a json response to the client
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	// Log the response write start
	logger := Log.WithFields(logrus.Fields{
		"action": "writeJSON",
		"status": status,
		"data":   data,
	})

	// Marshal the response data into JSON
	out, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal response data")
		return err
	}

	// Set custom headers if provided
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Write the JSON data to the response body
	_, err = w.Write(out)
	if err != nil {
		logger.WithError(err).Error("Failed to write response data")
		return err
	}

	logger.Info("Successfully wrote JSON response")
	return nil
}

// errorJSON takes an error, and optionally a response status code, and generates and sends
// a json error response
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	// Set the default status code to Bad Request
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	// Create the error response payload
	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	// Log the error
	logger := Log.WithFields(logrus.Fields{
		"action":     "errorJSON",
		"statusCode": statusCode,
		"error":      err.Error(),
	})

	// Write the error JSON response
	if writeErr := app.writeJSON(w, statusCode, payload); writeErr != nil {
		logger.WithError(writeErr).Error("Failed to write error JSON response")
		return writeErr
	}

	// Log the successful error response
	logger.Info("Successfully sent error JSON response")
	return nil
}
