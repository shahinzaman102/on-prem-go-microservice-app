package api

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Rabbit *amqp091.Connection
	Logger *logrus.Logger
}
