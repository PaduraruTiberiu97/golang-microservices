// Package main runs the UI that exercises broker-service test actions.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "test.page.gohtml")
	})

	fmt.Println("Starting front end service on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
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

	data.BrokerURL = os.Getenv("BROKER_URL")
	if data.BrokerURL == "" {
		data.BrokerURL = "http://localhost:8000"
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
