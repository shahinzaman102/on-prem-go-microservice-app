package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// SendMail is the handler for sending emails
func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	// Start timer to measure the duration of the email send request
	start := time.Now()

	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		// Use logrus logger
		app.Logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "SendMail",
			"status": "failure",
		}).Error("Failed to read JSON")

		// Increment error counter for mail send attempt
		MailSendErrors.Inc()
		MailSendAttempts.WithLabelValues("failure").Inc()
		// Record the failure duration
		MailSendDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		app.errorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		// Use logrus logger
		app.Logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"method": "SendMail",
			"status": "failure",
		}).Error("Failed to send email")

		// Increment error counter for mail send
		MailSendErrors.Inc()
		MailSendAttempts.WithLabelValues("failure").Inc()
		// Record the failure duration
		MailSendDuration.WithLabelValues("failure").Observe(time.Since(start).Seconds())
		app.errorJSON(w, err)
		return
	}

	// Increment success counter for mail send
	MailSendAttempts.WithLabelValues("success").Inc()
	// Record the success duration
	MailSendDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())

	payload := jsonResponse{
		Error:   false,
		Message: "sent to " + requestPayload.To,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
