// Package event provides RabbitMQ exchange and queue declaration helpers.
package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareLogsExchange(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // is this durable?
		false,        // auto-delete?
		false,        // internal?
		false,        // no-wait?
		nil,          // arguments
	)
}

func declareEphemeralQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    // name
		false, // durable?
		false, // delete when unused?
		true,  // exclusive?
		false, // no-wait?
		nil,   // arguments
	)
}
