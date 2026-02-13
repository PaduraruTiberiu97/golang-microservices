// Package main defines HTTP handlers for authentication operations.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (app *Config) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	if app.Repository == nil {
		_ = app.writeErrorJSON(w, errors.New("repository is not configured"))
		return
	}

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
	entry := struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}{
		Name: name,
		Data: data,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	logServiceURL := app.LoggerServiceURL
	if logServiceURL == "" {
		logServiceURL = defaultLoggerServiceURL
	}

	request, err := http.NewRequest(http.MethodPost, logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := app.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("logger service returned status %d", response.StatusCode)
	}

	return nil
}
