package handlers

import (
	"net/http"
	"testing"
)

func TestRegisterGetList(t *testing.T) {
	registry = map[string]http.HandlerFunc{}

	h1 := func(w http.ResponseWriter, r *http.Request) {}
	h2 := func(w http.ResponseWriter, r *http.Request) {}
	Register("BHandler", h2)
	Register("AHandler", h1)

	got, ok := Get("AHandler")
	if !ok || got == nil {
		t.Fatal("expected AHandler to be registered")
	}

	names := List()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "AHandler" || names[1] != "BHandler" {
		t.Fatalf("expected sorted names, got %v", names)
	}
}
