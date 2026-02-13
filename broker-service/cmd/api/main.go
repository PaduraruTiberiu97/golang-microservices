// Package main runs the broker API that orchestrates calls to downstream services.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const httpPort = "80"
const defaultRabbitMQURL = "amqp://guest:guest@rabbitmq"

type Config struct {
	Rabbit           *amqp.Connection
	HTTPClient       *http.Client
	AuthServiceURL   string
	MailServiceURL   string
	LoggerServiceURL string
	LoggerRPCAddr    string
	LoggerGRPCAddr   string
}

func main() {
	rabbitmqConn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatal("Could not connect to RabbitMQ. Exiting...", err)
	}
	defer rabbitmqConn.Close()

	app := Config{
		Rabbit:           rabbitmqConn,
		HTTPClient:       &http.Client{Timeout: 5 * time.Second},
		AuthServiceURL:   getenv("AUTH_SERVICE_URL", "http://authentication-service/authenticate"),
		MailServiceURL:   getenv("MAIL_SERVICE_URL", "http://mail-service/send"),
		LoggerServiceURL: getenv("LOGGER_SERVICE_URL", "http://logger-service/log"),
		LoggerRPCAddr:    getenv("LOGGER_RPC_ADDR", "logger-service:5001"),
		LoggerGRPCAddr:   getenv("LOGGER_GRPC_ADDR", "logger-service:50001"),
	}

	log.Printf("Starting broker service on port %s\n", httpPort)

	// define HTTP server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", httpPort),
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		log.Panic(err)
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
