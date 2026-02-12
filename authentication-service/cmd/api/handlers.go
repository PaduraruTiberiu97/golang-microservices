// Package main defines HTTP handlers for authentication operations.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.decodeJSON(w, r, &requestPayload)
	if err != nil {
		err := app.writeErrorJSON(w, err, http.StatusBadRequest)
		if err != nil {
			return
		}
		return
	}

	// validate the user
	user, err := app.Repository.GetByEmail(requestPayload.Email)
	if err != nil {
		err := app.writeErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		if err != nil {
			return
		}
		return
	}

	// validate password
	valid, err := app.Repository.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		err := app.writeErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		if err != nil {
			return
		}
		return
	}

	// log auth
	err = app.logAuthenticationEvent("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.writeErrorJSON(w, err)
		return
	}

	payload := JsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	if err = app.writeJSON(w, http.StatusAccepted, payload); err != nil {
		return
	}
}

func (app *Config) logAuthenticationEvent(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	_, err = app.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	return nil
}
