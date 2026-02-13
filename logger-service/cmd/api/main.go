// Package main runs logger-service over HTTP, RPC, and gRPC.
package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	httpPort = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	grpcPort = "50001"
)

var mongoClient *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	var err error

	// Connect to MongoDB.
	mongoClient, err = connectMongo()
	if err != nil {
		log.Panic(err)
	}

	// Create a context used when disconnecting from MongoDB.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()

	app := Config{Models: data.NewModels(mongoClient)}

	// Register the RPC server.
	if err = rpc.Register(&RPCServer{Models: app.Models}); err != nil {
		log.Panic(err)
	}

	go app.listenRPC()
	go app.listenGRPC()

	// Start HTTP server.
	log.Println("Starting logger-service on port", httpPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: app.routes(),
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func (app *Config) listenRPC() error {
	log.Println("Starting RPC server on port", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("Error accepting RPC connection:", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}

func connectMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connection, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Println("Error connecting to MongoDB", err)
		return nil, err
	}

	if err = connection.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")
	return connection, nil
}
