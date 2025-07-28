package api

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}

var (
	requestsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "broker_requests_total",
			Help: "Total number of requests processed by broker service.",
		},
		[]string{"action"},
	)

	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "broker_request_latency_seconds",
			Help:    "Histogram of request processing latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"action"},
	)

	requestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "broker_request_errors_total",
			Help: "Total number of failed requests by action.",
		},
		[]string{"action"},
	)

	rabbitFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "broker_rabbitmq_failures_total",
			Help: "Total number of RabbitMQ message failures.",
		},
	)

	grpcFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "broker_grpc_failures_total",
			Help: "Total number of failed gRPC requests.",
		},
	)

	rpcFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "broker_rpc_failures_total",
			Help: "Total number of failed RPC requests.",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsProcessed, requestLatency, requestErrors, rabbitFailures, grpcFailures, rpcFailures)
}

type RequestPayload struct {
	Action    string      `json:"action"`
	Auth      AuthPayload `json:"auth,omitempty"`
	LogRabbit LogPayload  `json:"logRabbit,omitempty"`
	LogRpc    LogPayload  `json:"logRpc,omitempty"`
	LogGrpc   LogPayload  `json:"logGrpc,omitempty"`
	Mail      MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func getTraceID(r *http.Request) string {
	return trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	payload := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	log.WithField("trace_id", traceID).Info("Broker hit successfully")
	_ = app.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	var requestPayload RequestPayload
	start := time.Now()

	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		requestErrors.WithLabelValues(requestPayload.Action).Inc()
		log.WithFields(logrus.Fields{
			"action":   requestPayload.Action,
			"error":    err.Error(),
			"trace_id": traceID,
		}).Error("Failed to read JSON payload")
		app.ErrorJSON(w, err)
		return
	}

	log.WithFields(logrus.Fields{
		"action":   requestPayload.Action,
		"payload":  requestPayload,
		"trace_id": traceID,
	}).Info("Received request")

	requestsProcessed.WithLabelValues(requestPayload.Action).Inc()

	defer func() {
		duration := time.Since(start).Seconds()
		requestLatency.WithLabelValues(requestPayload.Action).Observe(duration)

		log.WithFields(logrus.Fields{
			"action":   requestPayload.Action,
			"duration": duration,
			"status":   "success",
			"trace_id": traceID,
		}).Info("Request processed successfully")
	}()

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth, traceID)
	case "logRabbit":
		app.logEventViaRabbit(w, requestPayload.LogRabbit, traceID)
	case "logRpc":
		app.logItemViaRPC(w, requestPayload.LogRpc, traceID)
	case "mail":
		app.SendMail(w, requestPayload.Mail, traceID)
	default:
		requestErrors.WithLabelValues(requestPayload.Action).Inc()
		log.WithFields(logrus.Fields{"action": requestPayload.Action, "trace_id": traceID}).Error("Unknown action")
		app.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload, traceID string) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	log.WithFields(logrus.Fields{
		"email":    a.Email,
		"trace_id": traceID,
	}).Info("Sending authentication request")

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to create HTTP request")
		app.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to call authentication service")
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.WithFields(logrus.Fields{
		"status_code": response.StatusCode,
		"trace_id":    traceID,
	}).Info("Received response from authentication service")

	if response.StatusCode == http.StatusUnauthorized {
		log.WithField("trace_id", traceID).Warn("Invalid credentials")
		app.ErrorJSON(w, errors.New("invalid credentials"))
		return
	}

	if response.StatusCode >= 400 {
		requestErrors.WithLabelValues("auth").Inc()
		log.WithFields(logrus.Fields{"status": response.StatusCode, "trace_id": traceID}).Error("Auth service error")
		app.ErrorJSON(w, fmt.Errorf("auth service error (status: %d)", response.StatusCode))
		return
	}

	var jsonFromService JsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to decode auth response")
		app.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		requestErrors.WithLabelValues("auth").Inc()
		log.WithFields(logrus.Fields{"error": jsonFromService.Message, "trace_id": traceID}).Error("Authentication failed")
		app.ErrorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	log.WithField("trace_id", traceID).Info("Authentication successful")
	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload, traceID string) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	request, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		requestErrors.WithLabelValues("mail").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to create mail request")
		app.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		requestErrors.WithLabelValues("mail").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to call mail service")
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	log.WithFields(logrus.Fields{
		"status_code": response.StatusCode,
		"trace_id":    traceID,
	}).Info("Mail service response received")

	if response.StatusCode != http.StatusAccepted {
		requestErrors.WithLabelValues("mail").Inc()
		log.WithFields(logrus.Fields{"status_code": response.StatusCode, "trace_id": traceID}).Error("Error from mail service")
		app.ErrorJSON(w, errors.New("error calling mail service"))
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload, traceID string) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		rabbitFailures.Inc()
		requestErrors.WithLabelValues("logRabbit").Inc()
		log.WithFields(logrus.Fields{"name": l.Name, "data": l.Data, "error": err.Error(), "trace_id": traceID}).Error("Failed to push event to RabbitMQ")
		app.ErrorJSON(w, err)
		return
	}

	log.WithFields(logrus.Fields{
		"name":     l.Name,
		"data":     l.Data,
		"trace_id": traceID,
	}).Info("Event logged via RabbitMQ")

	var payload JsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.WriteJSON(w, http.StatusAccepted, payload)
}

// pushToQueue pushes a message into RabbitMQ
func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, err := json.MarshalIndent(&payload, "", "\t")
	if err != nil {
		return err
	}

	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload, traceID string) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		rpcFailures.Inc()
		requestErrors.WithLabelValues("logRpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to connect to RPC server")
		app.ErrorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload(l)

	var result string
	if err := client.Call("RPCServer.LogInfo", rpcPayload, &result); err != nil {
		rpcFailures.Inc()
		requestErrors.WithLabelValues("logRpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("RPC log failure")
		app.ErrorJSON(w, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: result,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	var requestPayload RequestPayload
	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to read gRPC request")
		app.ErrorJSON(w, err)
		return
	}

	conn, err := grpc.NewClient(
		"logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to connect to gRPC server")
		app.ErrorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.LogGrpc.Name,
			Data: requestPayload.LogGrpc.Data,
		},
	})
	if err != nil {
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error(), "trace_id": traceID}).Error("Failed to send log via gRPC")
		app.ErrorJSON(w, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Processed payload via gRPC"

	app.WriteJSON(w, http.StatusAccepted, payload)
}
