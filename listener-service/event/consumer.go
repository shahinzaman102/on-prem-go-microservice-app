package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var (
	// Define Prometheus metrics

	RabbitMessagesProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_processed_total",
			Help: "Total number of messages processed.",
		},
	)

	RabbitMessageProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "rabbitmq_message_processing_duration_seconds",
			Help:    "Histogram of message processing durations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	RabbitMessageErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rabbitmq_message_errors_total",
			Help: "Total number of message processing errors.",
		},
	)
)

// init function ensures metrics are registered only once
func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(RabbitMessagesProcessed)
	prometheus.MustRegister(RabbitMessageProcessingDuration)
	prometheus.MustRegister(RabbitMessageErrors)
}

type Consumer struct {
	conn *amqp.Connection
	log  *logrus.Logger
}

func NewConsumer(conn *amqp.Connection, log *logrus.Logger) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
		log:  log,
	}

	err := consumer.setup()
	if err != nil {
		log.WithError(err).Error("Failed to set up consumer")
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		consumer.log.WithError(err).Error("Failed to create channel")
		RabbitMessageErrors.Inc()
		return err
	}

	err = declareExchange(channel)
	if err != nil {
		consumer.log.WithError(err).Error("Failed to declare exchange")
		RabbitMessageErrors.Inc()
	}
	return err
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		consumer.log.WithError(err).Error("Failed to open channel")
		RabbitMessageErrors.Inc()
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		consumer.log.WithError(err).Error("Failed to declare queue")
		RabbitMessageErrors.Inc()
		return err
	}

	for _, s := range topics {
		err = ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			consumer.log.WithField("topic", s).WithError(err).Error("Failed to bind queue")
			RabbitMessageErrors.Inc()
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		consumer.log.WithError(err).Error("Failed to consume messages")
		RabbitMessageErrors.Inc()
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			err := json.Unmarshal(d.Body, &payload)
			if err != nil {
				consumer.log.WithError(err).Error("Failed to unmarshal message payload")
				RabbitMessageErrors.Inc()
				continue
			}

			start := time.Now()
			go handlePayload(payload, consumer.log)
			duration := time.Since(start).Seconds()
			RabbitMessageProcessingDuration.Observe(duration)

			consumer.log.WithFields(logrus.Fields{
				"message":  payload,
				"duration": duration,
			}).Info("Processed message")

			RabbitMessagesProcessed.Inc()
		}
	}()

	consumer.log.WithField("queue", q.Name).Info("Waiting for messages")
	<-forever

	return nil
}

func handlePayload(payload Payload, log *logrus.Logger) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload, log)
		if err != nil {
			log.WithError(err).Error("Failed to log event")
			RabbitMessageErrors.Inc()
		}
	default:
		log.Warn("Unhandled message type")
	}
}

func logEvent(entry Payload, log *logrus.Logger) error {
	jsonData, _ := json.Marshal(entry)

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		RabbitMessageErrors.Inc()
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		RabbitMessageErrors.Inc()
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		RabbitMessageErrors.Inc()
		return fmt.Errorf("failed to log event: status code %d", response.StatusCode)
	}

	log.WithFields(logrus.Fields{
		"payload": entry,
	}).Info("Event logged successfully")

	return nil
}
