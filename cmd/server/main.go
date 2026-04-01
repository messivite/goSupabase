package main

import (
	"log"
	"net/http"

	"github.com/messivite/gosupabase/config"
	_ "github.com/messivite/gosupabase/handlers"
	"github.com/messivite/gosupabase/server"
)

var (
	loadEnv        = config.LoadEnv
	newHTTPHandler = server.NewHandler
	listenAndServe = http.ListenAndServe
)

func run() error {
	cfg := loadEnv()
	handler := newHTTPHandler("api.yaml", cfg.SupabaseJWTSecret, cfg.SupabaseURL, cfg.SupabaseJWTValidationMode)

	addr := ":" + cfg.Port
	log.Printf("[gosupabase] server starting on %s", addr)
	return listenAndServe(addr, handler)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("[gosupabase] server error: %v", err)
	}
}
