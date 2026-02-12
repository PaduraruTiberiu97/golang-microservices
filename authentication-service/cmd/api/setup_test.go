// Package main provides shared test setup for authentication handlers.
package main

import (
	"authentication-service/data"
	"os"
	"testing"
)

var testApp Config

func TestMain(m *testing.M) {
	repo := data.NewPostgresTestRepository(nil)
	testApp.Repository = repo

	os.Exit(m.Run())
}
