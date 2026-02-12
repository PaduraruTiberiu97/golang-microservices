// Package event contains RabbitMQ publisher primitives used by the broker.
package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) ensureExchange() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return declareLogsExchange(channel)
}

func (e *Emitter) Publish(event string, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	log.Println("Pushing event:", event, "with severity:", severity)

	err = channel.Publish(
		"logs_topic",
		severity, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func NewEmitter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{connection: conn}
	err := emitter.ensureExchange()
	if err != nil {
		return Emitter{}, err
	}
	return emitter, nil
}
