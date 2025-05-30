package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

func init() {
	// Configure logger
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) setup() error {
	log.Info("Setting up the emitter")
	channel, err := e.connection.Channel()
	if err != nil {
		log.WithError(err).Error("Failed to create channel in emitter setup")
		return err
	}
	defer channel.Close()

	err = declareExchange(channel)
	if err != nil {
		log.WithError(err).Error("Failed to declare exchange in emitter setup")
		return err
	}

	log.Info("Emitter setup complete")
	return nil
}

func (e *Emitter) Push(event string, severity string) error {
	log.WithFields(logrus.Fields{
		"event":    event,
		"severity": severity,
	}).Info("Attempting to push event to channel")

	channel, err := e.connection.Channel()
	if err != nil {
		log.WithError(err).Error("Failed to open channel for publishing")
		return err
	}
	defer channel.Close()

	err = channel.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to publish message to channel")
		return err
	}

	log.WithFields(logrus.Fields{
		"event":    event,
		"severity": severity,
	}).Info("Successfully pushed event to channel")
	return nil
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	log.Info("Creating a new event emitter")

	emitter := Emitter{
		connection: conn,
	}

	err := emitter.setup()
	if err != nil {
		log.WithError(err).Error("Failed to set up event emitter")
		return Emitter{}, err
	}

	log.Info("Event emitter successfully created")
	return emitter, nil
}
