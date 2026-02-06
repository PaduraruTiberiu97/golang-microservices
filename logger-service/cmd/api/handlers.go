package main

import (
	"log-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// read JSON into var
	var requestPayload JSONPayload
	_ = app.ReadJSON(w, r, &requestPayload)

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		_ = app.errorJSON(w, err)
	}

	resp := JsonResponse{
		Error:   false,
		Message: "logged",
	}

	_ = app.writeJSON(w, http.StatusAccepted, resp)
}
