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

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitmqConn, err := connect()
	if err != nil {
		log.Fatal("Could not connect to RabbitMQ. Exiting...", err)
		os.Exit(1)
	}
	defer rabbitmqConn.Close()

	app := Config{
		Rabbit: rabbitmqConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	//define http server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Panic(err)
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
