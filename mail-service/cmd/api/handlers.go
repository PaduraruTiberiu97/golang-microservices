// Package main defines mail-service HTTP handlers.
package main

import "net/http"

func (app *Config) handleSendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := app.decodeJSON(w, r, &requestPayload)
	if err != nil {
		app.writeErrorJSON(w, err, http.StatusBadRequest)
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
		app.writeErrorJSON(w, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: "sent to " + requestPayload.To,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
