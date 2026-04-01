package api

import (
	"net/http"

	"github.com/messivite/gosupabase/config"
	_ "github.com/messivite/gosupabase/handlers"
	"github.com/messivite/gosupabase/server"
)

var handler http.Handler

func init() {
	cfg := config.LoadEnv()
	handler = server.NewHandler("api.yaml", cfg.SupabaseJWTSecret, cfg.SupabaseURL, cfg.SupabaseJWTValidationMode)
}

// Handler is the serverless entry point (e.g., Vercel).
func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
