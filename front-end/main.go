// Package main runs the UI that exercises broker-service test actions.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const defaultFrontendPort = "8081"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		renderTemplate(w, "test.page.gohtml")
	})

	frontendPort := frontendPortFromEnv()
	fmt.Printf("Starting front end service on port %s\n", frontendPort)
	server := &http.Server{
		Addr:              ":" + frontendPort,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func renderTemplate(w http.ResponseWriter, pageTemplate string) {
	partials := []string{
		"templates/base.layout.gohtml",
		"templates/header.partial.gohtml",
		"templates/footer.partial.gohtml",
	}

	templatePaths := []string{fmt.Sprintf("templates/%s", pageTemplate)}

	for _, partial := range partials {
		templatePaths = append(templatePaths, partial)
	}

	tmpl, err := template.ParseFiles(templatePaths...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data struct {
		BrokerURL string
	}

	data.BrokerURL = brokerURLFromEnv()

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func brokerURLFromEnv() string {
	if value := os.Getenv("BROKER_URL"); value != "" {
		return value
	}

	return "http://localhost:8000"
}

func frontendPortFromEnv() string {
	if value := os.Getenv("FRONTEND_PORT"); value != "" {
		return value
	}

	return defaultFrontendPort
}
