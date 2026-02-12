// Package main runs the broker API that orchestrates calls to downstream services.
package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const httpPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitmqConn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatal("Could not connect to RabbitMQ. Exiting...", err)
		os.Exit(1)
	}
	defer rabbitmqConn.Close()

	app := Config{
		Rabbit: rabbitmqConn,
	}

	log.Printf("Starting broker service on port %s\n", httpPort)

	// define HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: app.routes(),
	}

	if err = server.ListenAndServe(); err != nil {
		log.Panic(err)
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
