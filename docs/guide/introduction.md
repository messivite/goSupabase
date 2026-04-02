# Introduction

goSupaBase is a CLI toolkit and runtime library for building Go APIs backed by [Supabase](https://supabase.com/). It takes a **YAML-first** approach: you define your endpoints in `api.yaml`, then the CLI generates handler stubs, server wiring, and (optionally) deploy configs.

## What It Does

- **YAML endpoint definitions** — single `api.yaml` as the source of truth
- **Code generation** — handler stubs and server wiring from YAML
- **Runtime routing** — `api.yaml` loaded at startup, no regeneration needed for new endpoints
- **Supabase JWT auth** — HS256/ES256 (JWKS) with claims context and role guards
- **Configurable output** — custom directories via flags, `.gosupabase.yaml`, or `api.yaml`
- **Handlers-only mode** — generate stubs without local `server.go` (library consumers); full `gen` adds `server.go` only when local `middleware/` and `internal/yaml/` exist
- **Deploy scaffolding** — Vercel, Fly.io, Railway, and Render configs out of the box

## How It Works

**Full scaffold** (`gosupabase new`): handlers + optional local `server.go` (when `middleware/` and `internal/yaml/` exist).

**Library app** (`go get` + `gosupabase init`): `gosupabase gen --handlers-only` → `handlers/` only; routing lives in **`github.com/messivite/gosupabase/server`**.

```
api.yaml ──→ gosupabase gen ──→ handlers/  (+ local server.go only if deps exist)
                                     │
                                     ▼
                            go run ./cmd/server
                                     │
                                     ▼
                          api.yaml loaded at runtime
                          routes wired dynamically
```

1. Define endpoints in `api.yaml` (method, path, handler name, auth, roles).
2. Run `gosupabase gen` (full) or `gosupabase gen --handlers-only` (library consumers) to scaffold handler functions; full `gen` also writes local `server.go` only when the CLI detects the required local packages.
3. Implement your business logic in the generated handler files.
4. The server loads `api.yaml` at startup and wires routes through a handler registry — no code regeneration needed when you add new endpoints.

## Next Steps

- [Install goSupaBase](/guide/installation)
- [Quick Start — build your first API in 60 seconds](/guide/quick-start)
