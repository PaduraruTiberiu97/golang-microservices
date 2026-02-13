// Package main exposes logger-service write operations over gRPC.
package main

import (
	"context"
	"log"
	"log-service/data"
	"log-service/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

// Write persists a log entry received from gRPC clients.
func (l *LogServer) Write(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()
	if input == nil {
		return &logs.LogResponse{Result: "failed: empty payload"}, nil
	}

	// write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, nil
	}

	// return response
	res := &logs.LogResponse{Result: "success"}
	return res, nil
}

func (app *Config) listenGRPC() {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	logs.RegisterLogServiceServer(grpcServer, &LogServer{Models: app.Models})

	log.Printf("gRPC server started on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
