package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mustafaaksoy/gosupabase/handlers"
	yamlcfg "github.com/mustafaaksoy/gosupabase/internal/yaml"
	"github.com/mustafaaksoy/gosupabase/middleware"
)

// NewHandler loads api.yaml at runtime and builds a chi router with all
// registered handlers wired up, auth middleware applied where configured.
func NewHandler(apiPath, jwtSecret, supabaseURL, jwtValidationMode string) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	cfg, err := yamlcfg.Load(apiPath)
	if err != nil {
		log.Printf("[gosupabase] warning: could not load %s: %v", apiPath, err)
		return r
	}

	for _, ep := range cfg.Endpoints {
		h, ok := handlers.Get(ep.Handler)
		if !ok {
			log.Printf("[gosupabase] warning: handler %q not registered, skipping %s %s", ep.Handler, ep.Method, ep.Path)
			continue
		}

		fullPath := buildFullPath(cfg.BasePath, ep.Path)
		chiPath := yamlcfg.NormalizePath(fullPath)

		var chain http.Handler = h
		if len(ep.Roles) > 0 {
			chain = middleware.RequireRoles(ep.Roles...)(chain)
		}
		if ep.Auth {
			chain = middleware.SupabaseAuth(jwtSecret, supabaseURL, jwtValidationMode)(chain)
		}

		addRoute(r, ep.Method, chiPath, chain)
		log.Printf("[gosupabase] %s %s -> %s (auth=%v, roles=%v)", ep.Method, chiPath, ep.Handler, ep.Auth, ep.Roles)
	}

	return r
}

func buildFullPath(basePath, epPath string) string {
	base := strings.TrimSuffix(basePath, "/")
	path := epPath
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasPrefix(path, base) {
		return path
	}
	return base + path
}

func addRoute(r chi.Router, method, path string, h http.Handler) {
	switch strings.ToUpper(method) {
	case "GET":
		r.Get(path, h.ServeHTTP)
	case "POST":
		r.Post(path, h.ServeHTTP)
	case "PUT":
		r.Put(path, h.ServeHTTP)
	case "PATCH":
		r.Patch(path, h.ServeHTTP)
	case "DELETE":
		r.Delete(path, h.ServeHTTP)
	case "OPTIONS":
		r.Options(path, h.ServeHTTP)
	default:
		r.Method(method, path, h)
	}
}
