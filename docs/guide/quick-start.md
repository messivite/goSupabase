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

Use **`gosupabase gen --handlers-only`** so routing stays in the published [`server`](https://pkg.go.dev/github.com/messivite/gosupabase/server) package. **Full** `gosupabase gen` (without the flag) would write a local `server.go` that imports your module’s **`middleware/`** and **`internal/yaml/`** — those only exist in a full scaffold (e.g. `gosupabase new`). If they are missing, the CLI **refuses** to write `server.go` and tells you to use `--handlers-only`. Run `gen` from your module root so `go.mod` can be found.

## Hot-Reload Development

```bash
gosupabase dev
```

The dev server watches `.go` files and `api.yaml` — it auto-restarts on changes, so you get a tight edit-save-test loop.

## What Happens Under the Hood

1. **`gosupabase new`** creates directories, `go.mod`, `api.yaml`, `.env.example`, `middleware/`, `internal/yaml/`, and a `Makefile` — enough for **full** `gosupabase gen` later.
2. **`gosupabase gen`** reads `api.yaml` and always generates (or skips existing) handler files plus `handlers/registry.go`.
3. **Full** `gosupabase gen` (no `--handlers-only`) also writes `<serverDir>/server.go` **only if** `middleware/*.go` and `internal/yaml/*.go` exist under the module root and `go.mod` is reachable from the current directory.
4. **`gosupabase gen --handlers-only`** never writes `server.go` — you use `github.com/messivite/gosupabase/server` from `cmd/server` (e.g. after `gosupabase init`).
5. At startup, the running server loads `api.yaml`, resolves each handler by name from the registry, and attaches auth / role middleware as configured.
