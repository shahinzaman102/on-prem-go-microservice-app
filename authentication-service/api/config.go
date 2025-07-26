package api

import (
	"authentication/data"
	"database/sql"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB     *sql.DB
	Models data.Models
	Logger *logrus.Logger
	Redis  *redis.Client

	Metrics struct {
		RequestCount       *prometheus.CounterVec
		RequestLatency     *prometheus.HistogramVec
		ErrorCount         *prometheus.CounterVec
		PGConnectionStatus *prometheus.GaugeVec
	}
}
