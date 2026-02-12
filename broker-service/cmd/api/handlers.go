// Package main implements broker request handlers and service-to-service forwarding logic.
package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email string `json:"email"`
	Pass  string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) handleBroker(w http.ResponseWriter, r *http.Request) {
	payload := JsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) handleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.decodeJSON(w, r, &requestPayload)
	if err != nil {
		err := app.writeErrorJSON(w, err)
		if err != nil {
			return
		}
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.forwardAuthRequest(w, r, requestPayload.Auth)
	case "log":
		// app.logViaRabbitMQ(w, requestPayload.Log)
		app.logViaRPC(w, requestPayload.Log)
	case "mail":
		app.forwardMailRequest(w, requestPayload.Mail)
	default:
		err := app.writeErrorJSON(w, errors.New("invalid action"))
		if err != nil {
			return
		}
	}
}

func (app *Config) forwardMailRequest(w http.ResponseWriter, mail MailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	// call the mail service
	mailServiceURL := "http://mailer-service/send"

	// post to mail-service

	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_ = app.writeErrorJSON(w, fmt.Errorf("mail service returned status %d", response.StatusCode))
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Mail sent"

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) forwardLogRequestHTTP(w http.ResponseWriter, logPayload LogPayload) {
	jsonData, _ := json.MarshalIndent(logPayload, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_ = app.writeErrorJSON(w, fmt.Errorf("log service returned status %d", response.StatusCode))
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged"

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) forwardAuthRequest(w http.ResponseWriter, r *http.Request, authPayload AuthPayload) {
	// create some JSON we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(authPayload, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		_ = app.writeErrorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		_ = app.writeErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService JsonResponse

	//decode the JSON from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	// if we get an error message in jsonFromService.Error == true
	if jsonFromService.Error {
		_ = app.writeErrorJSON(w, errors.New(jsonFromService.Message))
		return
	}

	// everything worked, send the jsonFromService back to the caller
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) logViaRabbitMQ(w http.ResponseWriter, logPayload LogPayload) {
	err := app.publishLogEvent(logPayload.Name, logPayload.Data)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) publishLogEvent(name, msg string) error {
	emitter, err := event.NewEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	jsonData, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Publish(string(jsonData), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRPC(w http.ResponseWriter, logPayload LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	var rpcPayload RPCPayload
	rpcPayload.Name = logPayload.Name
	rpcPayload.Data = logPayload.Data

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = result

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.decodeJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(
		dialCtx,
		"logger-service:50001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	defer conn.Close()

	client := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Write(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})

	if err != nil {
		_ = app.writeErrorJSON(w, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged via GRPC"
	payload.Data = res.GetResult()

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
