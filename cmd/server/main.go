package main

import (
	"log"
	"net/http"

	"github.com/mustafaaksoy/gosupabase/config"
	_ "github.com/mustafaaksoy/gosupabase/handlers"
	"github.com/mustafaaksoy/gosupabase/server"
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
