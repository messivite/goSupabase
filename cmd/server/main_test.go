package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mustafaaksoy/gosupabase/config"
)

func TestRunStartsWithConfiguredAddr(t *testing.T) {
	prevLoad := loadEnv
	prevNew := newHTTPHandler
	prevListen := listenAndServe
	t.Cleanup(func() {
		loadEnv = prevLoad
		newHTTPHandler = prevNew
		listenAndServe = prevListen
	})

	loadEnv = func() *config.Config {
		return &config.Config{Port: "9999"}
	}
	newHTTPHandler = func(apiPath, jwtSecret, supabaseURL, jwtValidationMode string) http.Handler {
		return http.NewServeMux()
	}

	called := false
	listenAndServe = func(addr string, handler http.Handler) error {
		called = true
		if addr != ":9999" {
			t.Fatalf("addr = %s, want :9999", addr)
		}
		if handler == nil {
			t.Fatal("handler is nil")
		}
		return nil
	}

	if err := run(); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !called {
		t.Fatal("listenAndServe was not called")
	}
}

func TestRunReturnsServerError(t *testing.T) {
	prevLoad := loadEnv
	prevNew := newHTTPHandler
	prevListen := listenAndServe
	t.Cleanup(func() {
		loadEnv = prevLoad
		newHTTPHandler = prevNew
		listenAndServe = prevListen
	})

	loadEnv = func() *config.Config { return &config.Config{Port: "8080"} }
	newHTTPHandler = func(apiPath, jwtSecret, supabaseURL, jwtValidationMode string) http.Handler {
		return http.NewServeMux()
	}
	wantErr := errors.New("listen failed")
	listenAndServe = func(addr string, handler http.Handler) error {
		return wantErr
	}

	if err := run(); !errors.Is(err, wantErr) {
		t.Fatalf("run() error = %v, want %v", err, wantErr)
	}
}
