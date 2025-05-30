package api

import (
	"authentication/data"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB     *sql.DB
	Models data.Models
	Logger *logrus.Logger

	Metrics struct {
		RequestCount       *prometheus.CounterVec
		RequestLatency     *prometheus.HistogramVec
		ErrorCount         *prometheus.CounterVec
		PGConnectionStatus *prometheus.GaugeVec
	}
}
