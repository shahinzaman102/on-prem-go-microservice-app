package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// SendMail is the handler for sending emails
func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	// Start OpenTelemetry trace span
	start := time.Now()
	tracer := otel.Tracer("mailer-service")
	_, span := tracer.Start(r.Context(), "SendMailHandler")
	defer span.End()

	// Define request payload
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	// Parse request body
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.Logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "SendMail",
			"status": "failure",
		}).Error("Failed to read JSON")

		// Trace error
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to parse JSON")

		MailSendErrors.Inc()
		MailSendAttempts.WithLabelValues("failure").Inc()
		MailSendDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		app.errorJSON(w, err)
		return
	}

	// Add trace attributes
	span.SetAttributes(
		attribute.String("email.to", requestPayload.To),
		attribute.String("email.subject", requestPayload.Subject),
	)

	// Create message object
	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	// Send email
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.Logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "SendMail",
			"status": "failure",
		}).Error("Failed to send email")

		// Trace error
		span.RecordError(err)
		span.SetStatus(codes.Error, "SMTP send failed")

		MailSendErrors.Inc()
		MailSendAttempts.WithLabelValues("failure").Inc()
		MailSendDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		app.errorJSON(w, err)
		return
	}

	// Record successful metrics
	MailSendAttempts.WithLabelValues("success").Inc()
	MailSendDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())

	// Respond
	payload := jsonResponse{
		Error:   false,
		Message: "sent to " + requestPayload.To,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
