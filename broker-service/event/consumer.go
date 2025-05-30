package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// Logger instance
var log = logrus.New()

func init() {
	// Configure logger
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	log.Info("Creating a new consumer")

	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		log.WithError(err).Error("Failed to setup consumer")
		return Consumer{}, err
	}

	log.Info("Consumer successfully created")
	return consumer, nil
}

func (consumer *Consumer) setup() error {
	log.Info("Setting up the consumer")
	channel, err := consumer.conn.Channel()
	if err != nil {
		log.WithError(err).Error("Failed to create channel")
		return err
	}

	err = declareExchange(channel)
	if err != nil {
		log.WithError(err).Error("Failed to declare exchange")
		return err
	}

	log.Info("Consumer setup complete")
	return nil
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	log.WithField("topics", topics).Info("Consumer is starting to listen to topics")
	ch, err := consumer.conn.Channel()
	if err != nil {
		log.WithError(err).Error("Failed to open channel")
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		log.WithError(err).Error("Failed to declare random queue")
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
			log.WithError(err).WithField("topic", s).Error("Failed to bind queue to topic")
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Error("Failed to consume messages from queue")
		return err
	}

	log.WithField("queue", q.Name).Info("Consumer is waiting for messages")
	forever := make(chan bool)

	go func() {
		for d := range messages {
			var payload Payload
			err := json.Unmarshal(d.Body, &payload)
			if err != nil {
				log.WithError(err).Error("Failed to unmarshal message")
				continue
			}

			log.WithField("payload", payload).Info("Received a message")
			go handlePayload(payload)
		}
	}()

	<-forever
	return nil
}

func handlePayload(payload Payload) {
	log.WithField("payload", payload).Info("Handling payload")
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.WithError(err).WithField("payload", payload).Error("Failed to log event")
		}
	case "auth":
		log.WithField("payload", payload).Info("Handling authentication payload")
		// Additional authentication logic here
	default:
		log.WithField("payload", payload).Warning("Unknown payload type")
		err := logEvent(payload)
		if err != nil {
			log.WithError(err).WithField("payload", payload).Error("Failed to log event")
		}
	}
}

func logEvent(entry Payload) error {
	log.WithField("entry", entry).Info("Logging event")
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.WithError(err).Error("Failed to create HTTP request for log service")
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("Failed to send HTTP request to log service")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.WithField("status_code", response.StatusCode).Error("Log service responded with unexpected status")
		return fmt.Errorf("unexpected status code from log service: %d", response.StatusCode)
	}

	log.WithField("entry", entry).Info("Event successfully logged")
	return nil
}
