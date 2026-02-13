// Package main starts the mail HTTP API and wires SMTP configuration from env vars.
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Mailer Mail
}

const httpPort = "80"

func main() {
	app := Config{
		Mailer: mailerFromEnv(),
	}

	log.Println("Starting mail service on port ", httpPort)

	srv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func mailerFromEnv() Mail {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		port = 25
	}

	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_NAME"),
		FromAddress: os.Getenv("MAIL_ADDRESS"),
	}

	return m
}
