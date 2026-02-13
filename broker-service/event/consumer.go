// Package event implements RabbitMQ consumption and forwarding to logger-service.
package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
}

var consumerHTTPClient = &http.Client{Timeout: 5 * time.Second}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{conn: conn}

	err := consumer.ensureExchange()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) ensureExchange() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return declareLogsExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func loggerServiceURL() string {
	if url := os.Getenv("LOGGER_SERVICE_URL"); url != "" {
		return url
	}

	return "http://logger-service/log"
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queue, err := declareEphemeralQueue(ch)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		err = ch.QueueBind(
			queue.Name,   // queue name
			topic,        // routing key
			"logs_topic", // exchange
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	log.Printf("Waiting for messages on exchange [Exchange, Queue] [logs_topic, %s]", queue.Name)
	for d := range messages {
		var payload Payload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Printf("Error decoding queue payload: %v", err)
			continue
		}

		go dispatchPayload(payload)
	}

	return errors.New("message channel closed")
}

func dispatchPayload(payload Payload) {
	if payload.Name == "auth" {
		return
	}

	if err := forwardLogEvent(payload); err != nil {
		log.Printf("Error handling payload: %v", err)
	}
}

func forwardLogEvent(entry Payload) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, loggerServiceURL(), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := consumerHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("logger service returned status %d", response.StatusCode)
	}

	return nil
}
