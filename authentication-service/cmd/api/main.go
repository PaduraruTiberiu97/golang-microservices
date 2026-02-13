// Package main boots the authentication HTTP service and wires its data dependencies.
package main

import (
	"authentication-service/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const httpPort = "80"

type Config struct {
	Repository data.Repository
	HTTPClient *http.Client
}

func main() {
	log.Println("Starting authentication service")

	conn := connectToPostgres()
	if conn == nil {
		log.Panic("Can't connect to database")
	}

	app := Config{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	app.setupRepository(conn)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func openPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToPostgres() *sql.DB {
	dsn := os.Getenv("DSN")
	var attempts int64

	for {
		connection, err := openPostgresDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			attempts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if attempts > 10 {
			log.Println("Too many postgres connections...")
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (app *Config) setupRepository(conn *sql.DB) {
	app.Repository = data.NewPostgresRepository(conn)
}
