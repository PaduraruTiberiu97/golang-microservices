// Package main defines mail-service HTTP handlers.
package main

import (
	"errors"
	"net/http"
	"strings"
)

type SendMailRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) handleSendMail(w http.ResponseWriter, r *http.Request) {
	var requestPayload SendMailRequest

	err := app.decodeJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.writeErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(requestPayload.To) == "" {
		_ = app.writeErrorJSON(w, errors.New("to is required"), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(requestPayload.Subject) == "" {
		_ = app.writeErrorJSON(w, errors.New("subject is required"), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(requestPayload.Message) == "" {
		_ = app.writeErrorJSON(w, errors.New("message is required"), http.StatusBadRequest)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "sent to " + requestPayload.To,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
