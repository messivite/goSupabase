package main

import (
	"testing"

	"github.com/mustafaaksoy/gosupabase/config"
	"github.com/mustafaaksoy/gosupabase/server"
)

func TestServerDependenciesCanInitialize(t *testing.T) {
	cfg := config.LoadEnv()
	h := server.NewHandler("api.yaml", cfg.SupabaseJWTSecret, cfg.SupabaseURL, cfg.SupabaseJWTValidationMode)
	if h == nil {
		t.Fatal("expected handler to be initialized")
	}
}
