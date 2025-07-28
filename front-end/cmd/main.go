package main

import (
	"context"
	"embed"
	"fmt"
	"frontend/internal/tracing"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestSize)
	prometheus.MustRegister(httpResponseSize)
}

func main() {
	ctx := context.Background()

	// Initialize tracing
	shutdown, err := tracing.InitTracer(ctx, "front-end")
	if err != nil {
		log.Panic(err)
	}
	defer shutdown(ctx)

	// Register HTTP routes with otelhttp instrumentation
	http.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(healthz), "healthz"))
	http.Handle("/ready", otelhttp.NewHandler(http.HandlerFunc(ready), "ready"))
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		reqSize, _ := io.Copy(io.Discard, r.Body)

		log.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("Received request")

		span.AddEvent("render.started")

		status := render(ctx, w, "test.page.gohtml")

		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, strconv.Itoa(status), r.URL.Path).Inc()
		httpRequestDuration.WithLabelValues(r.Method, strconv.Itoa(status), r.URL.Path).Observe(duration)
		httpRequestSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(reqSize))

		log.WithFields(logrus.Fields{
			"method":    r.Method,
			"status":    status,
			"duration":  duration,
			"timestamp": time.Now(),
		}).Info("Request handled successfully")

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.Int("http.status_code", status),
		)
	}), "Handle /"))

	log.Info("Starting front-end service on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Liveness probe passed: /healthz")
}

func ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Info("Readiness probe passed: /ready")
}

var tc = make(map[string]*template.Template)

//go:embed templates
var templateFS embed.FS

func render(ctx context.Context, w http.ResponseWriter, t string) int {
	var tmpl *template.Template
	var err error

	tr := otel.Tracer("front-end")
	_, span := tr.Start(ctx, "RenderTemplate:"+t)
	defer span.End()

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

	rec := &responseRecorder{ResponseWriter: w}
	if err := tmpl.Execute(rec, data); err != nil {
		log.WithError(err).Error("Failed to execute template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	httpResponseSize.WithLabelValues("GET", strconv.Itoa(http.StatusOK), t).Observe(float64(rec.size))
	return http.StatusOK
}

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

type responseRecorder struct {
	http.ResponseWriter
	size int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}
