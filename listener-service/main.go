// Package main starts a RabbitMQ consumer that forwards events to logger-service.
package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// try to connect to RabbitMQ
	rabbitmqConn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatal("Could not connect to RabbitMQ. Exiting...", err)
		os.Exit(1)
	}
	defer rabbitmqConn.Close()
	// start listening to messages
	log.Println("Listening and consuming RabbitMQ messages...")

	// create consumer
	consumer, err := event.NewConsumer(rabbitmqConn)
	if err != nil {
		panic(err)
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

	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMq not yet ready...")
			attempts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = conn
			break
		}

		if attempts > 5 {
			fmt.Println("Too many attempts to connect to RabbitMQ. Exiting...", err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(attempts), 2)) * time.Second
		log.Println("Backing off for", backoff)
		time.Sleep(backoff)

		continue
	}

	return connection, nil
}
