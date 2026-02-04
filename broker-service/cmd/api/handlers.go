package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	// Mail   Mail   `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email string `json:"email"`
	Pass  string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPaylod RequestPayload

	err := app.ReadJSON(w, r, &requestPaylod)
	if err != nil {
		err := app.errorJSON(w, err)
		if err != nil {
			return
		}
		return
	}

	switch requestPaylod.Action {
	case "auth":
		app.authenticate(w, r, requestPaylod.Auth)
	default:
		err := app.errorJSON(w, errors.New("Invalid action"))
		if err != nil {
			return
		}
	}
}

func (app *Config) authenticate(w http.ResponseWriter, r *http.Request, authPayload AuthPayload) {
	// create some JSON we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(authPayload, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusUnauthorized {
		_ = app.errorJSON(w, errors.New("Invalid credentials"))
	} else if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService JsonResponse

	//decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	// if we get an error message in jsonFromService.Error == true
	if jsonFromService.Error {
		_ = app.errorJSON(w, errors.New(jsonFromService.Message))
		return
	}

	// everything worked, send the jsonFromService back to the caller
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	_ = app.writeJSON(w, http.StatusOK, payload)
}
