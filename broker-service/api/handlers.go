package api

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logrus.New()

func init() {
	// Set log format to JSON, which is suitable for structured logging
	log.SetFormatter(&logrus.JSONFormatter{})
	// Optionally, set log level
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

// RequestPayload describes the JSON that this service accepts as an HTTP Post request
type RequestPayload struct {
	Action    string      `json:"action"`
	Auth      AuthPayload `json:"auth,omitempty"`
	LogRabbit LogPayload  `json:"logRabbit,omitempty"`
	LogRpc    LogPayload  `json:"logRpc,omitempty"`
	LogGrpc   LogPayload  `json:"logGrpc,omitempty"`
	Mail      MailPayload `json:"mail,omitempty"`
}

// MailPayload is the embedded type (in RequestPayload) that describes an email message to be sent
type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// AuthPayload is the embedded type (in RequestPayload) that describes an authentication request
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogPayload is the embedded type (in RequestPayload) that describes a request to log something
type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Broker is a test handler, just to make sure we can hit the broker from a web client
func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	log.Info("Broker hit successfully")
	_ = app.WriteJSON(w, http.StatusOK, payload)
}

// HandleSubmission is the main point of entry into the broker. It accepts a JSON
// payload and performs an action based on the value of "action" in that JSON.
func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	start := time.Now() // Start measuring time

	// Log the incoming request payload
	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		requestErrors.WithLabelValues(requestPayload.Action).Inc() // Increment error count for this action
		log.WithFields(logrus.Fields{"action": requestPayload.Action, "error": err.Error()}).Error("Failed to read JSON payload")
		app.ErrorJSON(w, err)
		return
	}

	// Log the request payload for debugging purposes
	log.WithFields(logrus.Fields{
		"action":  requestPayload.Action,
		"payload": requestPayload,
	}).Info("Received request")

	// Record request for Prometheus
	requestsProcessed.WithLabelValues(requestPayload.Action).Inc()

	// Use defer to record latency after processing
	defer func() {
		duration := time.Since(start).Seconds()
		requestLatency.WithLabelValues(requestPayload.Action).Observe(duration)

		log.WithFields(logrus.Fields{
			"action":   requestPayload.Action,
			"duration": duration,
			"status":   "success",
		}).Info("Request processed successfully")
	}()

	// Handle different actions based on the payload
	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "logRabbit":
		app.logEventViaRabbit(w, requestPayload.LogRabbit)
	case "logRpc":
		app.logItemViaRPC(w, requestPayload.LogRpc)
	// case "logGrpc":
	// 	app.LogViaGRPC(w, requestPayload.LogGrpc)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		requestErrors.WithLabelValues(requestPayload.Action).Inc() // Increment error count for unknown action
		log.WithFields(logrus.Fields{"action": requestPayload.Action}).Error("Unknown action")
		app.ErrorJSON(w, errors.New("unknown action"))
	}
}

// authenticate calls the authentication microservice and sends back the appropriate response
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// log the request
	log.WithFields(logrus.Fields{
		"email": a.Email,
	}).Info("Sending authentication request")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc() // Increment error count for auth action
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create HTTP request")
		app.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc() // Increment error count for auth action
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to call authentication service")
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// Log the response status code
	log.WithFields(logrus.Fields{
		"status_code": response.StatusCode,
	}).Info("Received response from authentication service")

	if response.StatusCode == http.StatusUnauthorized {
		log.WithFields(logrus.Fields{
			"status": response.StatusCode,
		}).Warn("Invalid credentials")
		app.ErrorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		requestErrors.WithLabelValues("auth").Inc() // Increment error count for auth action
		log.WithFields(logrus.Fields{
			"status": response.StatusCode,
		}).Error("Error calling authentication service")
		app.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService JsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		requestErrors.WithLabelValues("auth").Inc() // Increment error count for auth action
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to decode response from authentication service")
		app.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		requestErrors.WithLabelValues("auth").Inc() // Increment error count for auth action
		log.WithFields(logrus.Fields{
			"error": jsonFromService.Message,
		}).Error("Authentication failed")
		app.ErrorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	log.WithFields(logrus.Fields{
		"email": a.Email,
	}).Info("Authentication successful")

	app.WriteJSON(w, http.StatusAccepted, payload)
}

// SendMail sends email by calling the mail microservice
func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	// call the mail service
	mailServiceURL := "http://mailer-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		requestErrors.WithLabelValues("mail").Inc() // Increment error count for mail action
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to create mail request")
		app.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		requestErrors.WithLabelValues("mail").Inc() // Increment error count for mail action
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to call mail service")
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// Log the response status code
	log.WithFields(logrus.Fields{
		"status_code": response.StatusCode,
	}).Info("Received response from mail service")

	// make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		requestErrors.WithLabelValues("mail").Inc() // Increment error count for mail action
		log.WithFields(logrus.Fields{
			"status_code": response.StatusCode,
		}).Error("Error calling mail service")
		app.ErrorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back json
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.WriteJSON(w, http.StatusAccepted, payload)
}

// logEventViaRabbit logs an event using the logger-service. It makes the call by pushing the data to RabbitMQ.
func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		rabbitFailures.Inc()
		requestErrors.WithLabelValues("logRabbit").Inc() // Increment error count for logRabbit action
		log.WithFields(logrus.Fields{"name": l.Name, "data": l.Data, "error": err.Error()}).Error("Failed to push event to RabbitMQ")
		app.ErrorJSON(w, err)
		return
	}

	log.WithFields(logrus.Fields{
		"name": l.Name,
		"data": l.Data,
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

// logItemViaRPC logs an item by making an RPC call to the logger microservice
func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		rpcFailures.Inc()
		requestErrors.WithLabelValues("logRpc").Inc() // Increment error count for logRpc action
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to connect to RPC server")
		app.ErrorJSON(w, err)
		return
	}

	// Type conversion
	rpcPayload := RPCPayload(l)

	var result string
	if err := client.Call("RPCServer.LogInfo", rpcPayload, &result); err != nil {
		rpcFailures.Inc()
		requestErrors.WithLabelValues("logRpc").Inc() // Increment error count for logRpc action
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to log via RPC")
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
	var requestPayload RequestPayload

	// Read the incoming JSON payload
	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		// Increment error count for logGrpc action
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to read gRPC request")
		app.ErrorJSON(w, err)
		return
	}

	// Establish a gRPC connection using the new client approach
	conn, err := grpc.NewClient(
		"logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		// Increment error count for logGrpc action
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to connect to gRPC server")
		app.ErrorJSON(w, err)
		return
	}
	defer conn.Close()

	// Create the client and call WriteLog
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
		// Increment error count for logGrpc action
		grpcFailures.Inc()
		requestErrors.WithLabelValues("logGrpc").Inc()
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to send log via gRPC")
		app.ErrorJSON(w, err)
		return
	}

	// Success response
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Processed payload via gRPC"

	app.WriteJSON(w, http.StatusAccepted, payload)
}
