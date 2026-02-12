// Package main defines logger-service HTTP handlers.
package main

import (
	"log"
	"log-service/data"
	"net/http"
)

type LogRequestPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) handleWriteLog(w http.ResponseWriter, r *http.Request) {
	// read JSON into var
	var requestPayload LogRequestPayload
	if err := app.decodeJSON(w, r, &requestPayload); err != nil {
		_ = app.writeErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		log.Println("Error trying to insert", err)
		return
	}

	resp := JsonResponse{
		Error:   false,
		Message: "logged",
	}

	_ = app.writeJSON(w, http.StatusAccepted, resp)
}
