package main

import (
	"log"
	"net/http"

	"github.com/mustafaaksoy/gosupabase/config"
	_ "github.com/mustafaaksoy/gosupabase/handlers"
	"github.com/mustafaaksoy/gosupabase/server"
)

func main() {
	cfg := config.LoadEnv()

	handler := server.NewHandler("api.yaml", cfg.SupabaseJWTSecret, cfg.SupabaseURL, cfg.SupabaseJWTValidationMode)

	addr := ":" + cfg.Port
	log.Printf("[gosupabase] server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("[gosupabase] server error: %v", err)
	}
}
