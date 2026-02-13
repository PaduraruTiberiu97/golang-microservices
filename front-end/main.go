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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		renderTemplate(w, "test.page.gohtml")
	})

	fmt.Println("Starting front end service on port 8081")
	server := &http.Server{
		Addr:              ":8081",
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
