package main

import (
	"context"
	"log"
	"log-service/data"
	"time"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (r *RPCServer) LogInfo(payload RPCPayload, response *string) error {
	colelction := client.Database("logs").Collection("logs")
	_, err := colelction.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("error writing to Mongo: ", err)
		return err
	}

	*response = "Processed payload via RPC: " + payload.Name
	return nil

}
