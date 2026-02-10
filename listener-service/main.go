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
	// try to connect to rabbitmq
	rabbitmqConn, err := connect()
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

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMq not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = conn
			break
		}

		if counts > 5 {
			fmt.Println("Too many attempts to connect to RabbitMQ. Exiting...", err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off for", backOff)
		time.Sleep(backOff)

		continue
	}

	return connection, nil
}
