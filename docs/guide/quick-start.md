# Quick Start

## New Project

```bash
go install github.com/messivite/gosupabase/cmd/gosupabase@latest

gosupabase new my-api
cd my-api
go mod tidy
gosupabase gen
go run ./cmd/server
```

This scaffolds a complete project structure with `api.yaml`, handler stubs, a server entry point, and a `Makefile`.

## Existing Project

You need **`go.mod`** first (`go mod init <module>`).

```bash
gosupabase init          # api.yaml, .env.example, cmd/server/main.go (library server)
go get github.com/messivite/gosupabase@latest && go mod tidy
gosupabase setup         # .env + .gosupabase.yaml
gosupabase add endpoint "POST /tracks" --auth
gosupabase gen --handlers-only
go run ./cmd/server
```

Use **`gen --handlers-only`** so routing stays in the published `server` package; full `gen` also writes a local `server/server.go` that expects vendored `middleware` (advanced).

## Hot-Reload Development

```bash
gosupabase dev
```

The dev server watches `.go` files and `api.yaml` — it auto-restarts on changes, so you get a tight edit-save-test loop.

## What Happens Under the Hood

1. `gosupabase new` creates directories, `go.mod`, `api.yaml`, `.env.example`, and a `Makefile`.
2. `gosupabase gen` reads `api.yaml` and generates:
   - `handlers/<name>.go` — one file per endpoint with an `init()` that registers the handler.
   - `handlers/registry.go` — the handler lookup map.
   - `server/server.go` — chi router setup that loads `api.yaml` and wires routes at runtime.
3. At startup, the server reads `api.yaml`, resolves each handler by name from the registry, and attaches auth / role middleware as configured.
