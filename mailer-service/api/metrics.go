package api

import "github.com/prometheus/client_golang/prometheus"

var (
	// Track the number of email send attempts
	MailSendAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mailer_service_mail_send_attempts_total",
			Help: "Total number of email send attempts.",
		},
		[]string{"status"}, // Status label to track success/failure
	)

	// Track the number of mail send errors
	MailSendErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailer_service_mail_send_errors_total",
			Help: "Total number of errors occurred while sending emails.",
		},
	)

	// Track the duration of the mail send operation
	MailSendDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mailer_service_mail_send_duration_seconds",
			Help:    "Histogram of mail send durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"}, // You can add status label to track success/failure
	)
)

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(MailSendAttempts)
	prometheus.MustRegister(MailSendErrors)
	prometheus.MustRegister(MailSendDuration)
}
