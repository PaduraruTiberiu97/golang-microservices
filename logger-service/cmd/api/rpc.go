// Package main exposes logger-service write operations over net/rpc.
package main

import (
	"log"
	"log-service/data"
)

type RPCServer struct {
	Models data.Models
}

type RPCPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// LogInfo is called remotely by broker-service to persist a log entry.
func (r *RPCServer) LogInfo(payload RPCPayload, response *string) error {
	err := r.Models.LogEntry.Insert(data.LogEntry{
		Name: payload.Name,
		Data: payload.Data,
	})
	if err != nil {
		log.Println("error writing to Mongo: ", err)
		return err
	}

	*response = "Processed payload via RPC: " + payload.Name
	return nil
}
