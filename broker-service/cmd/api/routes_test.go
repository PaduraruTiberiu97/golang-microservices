// Package main verifies broker HTTP route wiring.
package main

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRoutesExist(t *testing.T) {
	app := Config{}

	routes := app.routes().(chi.Router)

	assertRouteExists(t, routes, "/")
	assertRouteExists(t, routes, "/handle")
	assertRouteExists(t, routes, "/log-grpc")
}

func assertRouteExists(t *testing.T, routes chi.Router, expectedRoute string) {
	t.Helper()

	found := false
	_ = chi.Walk(routes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == expectedRoute {
			found = true
		}
		return nil
	})

	if !found {
		t.Fatalf("route %s was not registered", expectedRoute)
	}
}
