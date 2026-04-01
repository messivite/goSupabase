package scaffold

const handlerTemplate = `package handlers

import (
	"encoding/json"
	"net/http"
)

func init() {
	Register("{{.Name}}", {{.Name}})
}

func {{.Name}}(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handler": "{{.Name}}",
		"status":  "ok",
	})
}
`

const registryTemplate = `package handlers

import (
	"net/http"
	"sort"
)

var registry = map[string]http.HandlerFunc{}

func Register(name string, h http.HandlerFunc) {
	registry[name] = h
}

func Get(name string) (http.HandlerFunc, bool) {
	h, ok := registry[name]
	return h, ok
}

func List() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
`

const serverTemplate = `package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"{{.Module}}/handlers"
	yamlcfg "{{.Module}}/internal/yaml"
	"{{.Module}}/middleware"
)

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
`

const mainServerTemplate = `package main

import (
	"log"
	"net/http"

	"{{.Module}}/config"
	_ "{{.Module}}/handlers"
	"{{.Module}}/server"
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
`

const makefileTemplate = `.PHONY: build run gen tidy

build:
	go build ./...

run:
	go run ./cmd/server

gen:
	go run ./cmd/gosupabase gen

tidy:
	go mod tidy
`

const apiYAMLTemplate = `version: "1"
basePath: /api
output:
  serverDir: server
  handlersDir: handlers

endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false
`
