// Package main starts a RabbitMQ consumer that forwards events to logger-service.
package main

import (
	"listener/event"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultRabbitMQURL = "amqp://guest:guest@rabbitmq"

func main() {
	// try to connect to RabbitMQ
	rabbitmqConn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatal("Could not connect to RabbitMQ. Exiting...", err)
	}
	defer rabbitmqConn.Close()
	// start listening to messages
	log.Println("Listening and consuming RabbitMQ messages...")

	// create consumer
	consumer, err := event.NewConsumer(rabbitmqConn)
	if err != nil {
		log.Fatal(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println("Error listening for messages", err)
	}
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	var attempts int64
	backoff := 1 * time.Second
	var connection *amqp.Connection
	rabbitMQURL := getenv("RABBITMQ_URL", defaultRabbitMQURL)

	for {
		conn, err := amqp.Dial(rabbitMQURL)
		if err != nil {
			log.Println("RabbitMQ not yet ready...")
			attempts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = conn
			break
		}

		if attempts > 5 {
			log.Println("Too many attempts to connect to RabbitMQ. Exiting...", err)
			return nil, err
		}

		backoff = time.Duration(attempts*attempts) * time.Second
		log.Println("Backing off for", backoff)
		time.Sleep(backoff)
	}

	return connection, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
