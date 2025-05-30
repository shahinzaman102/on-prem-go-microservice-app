package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	// Prometheus metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests received.",
		},
		[]string{"method", "status", "path"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status", "path"},
	)
	httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Histogram of request sizes.",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"method", "path"},
	)
	httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Histogram of response sizes.",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"method", "status", "path"},
	)

	log = logrus.New()
)

func init() {
	// Configure logger
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestSize)
	prometheus.MustRegister(httpResponseSize)
}

func main() {
	// Health check endpoints for liveness and readiness probes
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Main app handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Measure request size
		reqSize, _ := io.Copy(io.Discard, r.Body)

		// Log request
		log.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("Received request")

		// Simulate serving the template
		status := render(w, "test.page.gohtml")

		// Measure response time
		duration := time.Since(start).Seconds()

		// Record metrics
		httpRequestsTotal.WithLabelValues(r.Method, strconv.Itoa(status), r.URL.Path).Inc()
		httpRequestDuration.WithLabelValues(r.Method, strconv.Itoa(status), r.URL.Path).Observe(duration)
		httpRequestSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(reqSize))

		log.WithFields(logrus.Fields{
			"method":    r.Method,
			"status":    status,
			"duration":  duration,
			"timestamp": time.Now(),
		}).Info("Request handled successfully")
	})

	log.Info("Starting front-end service on port 8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// Healthz (Liveness) endpoint
func healthz(w http.ResponseWriter, r *http.Request) {
	// Here we return a 200 OK if the service is alive
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Liveness probe passed: /healthz")
}

// Ready (Readiness) endpoint
func ready(w http.ResponseWriter, r *http.Request) {
	// Here we return a 200 OK if the service is ready to serve traffic
	// You could add additional checks (like DB connection or any external services) if needed
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Readiness probe passed: /ready")
}

// Template cache
var tc = make(map[string]*template.Template)

//go:embed templates
var templateFS embed.FS

func render(w http.ResponseWriter, t string) int {
	var tmpl *template.Template
	var err error

	// Check template cache
	if _, exists := tc[t]; !exists {
		log.WithField("template", t).Info("Parsing and caching template")
		err = createTemplateCache(t)
		if err != nil {
			log.WithError(err).Error("Failed to create template cache")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return http.StatusInternalServerError
		}
	} else {
		log.WithField("template", t).Info("Using cached template")
	}

	tmpl = tc[t]
	data := struct {
		BrokerURL string
	}{
		BrokerURL: os.Getenv("BROKER_URL"),
	}

	// Capture response size
	rec := &responseRecorder{ResponseWriter: w}
	if err := tmpl.Execute(rec, data); err != nil {
		log.WithError(err).Error("Failed to execute template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	// Measure response size
	httpResponseSize.WithLabelValues("GET", strconv.Itoa(http.StatusOK), t).Observe(float64(rec.size))

	return http.StatusOK
}

// Create template cache
func createTemplateCache(t string) error {
	templates := []string{
		fmt.Sprintf("templates/%s", t),
		"templates/header.partial.gohtml",
		"templates/footer.partial.gohtml",
		"templates/base.layout.gohtml",
	}

	tmpl, err := template.ParseFS(templateFS, templates...)
	if err != nil {
		log.WithError(err).Error("Failed to parse template files")
		return err
	}

	tc[t] = tmpl
	return nil
}

// Response recorder to measure response size
type responseRecorder struct {
	http.ResponseWriter
	size int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}
