# Introduction

goSupaBase is a CLI toolkit and runtime library for building Go APIs backed by [Supabase](https://supabase.com/). It takes a **YAML-first** approach: you define your endpoints in `api.yaml`, then the CLI generates handler stubs, server wiring, and (optionally) deploy configs.

## What It Does

- **YAML endpoint definitions** — single `api.yaml` as the source of truth
- **Code generation** — handler stubs and server wiring from YAML
- **Runtime routing** — `api.yaml` loaded at startup, no regeneration needed for new endpoints
- **Supabase JWT auth** — HS256/ES256 (JWKS) with claims context and role guards
- **Configurable output** — custom directories via flags, `.gosupabase.yaml`, or `api.yaml`
- **Handlers-only mode** — generate stubs without touching the server
- **Deploy scaffolding** — Vercel, Fly.io, Railway, and Render configs out of the box

## How It Works

```
api.yaml ──→ gosupabase gen ──→ handlers/ + server/
                                     │
                                     ▼
                            go run ./cmd/server
                                     │
                                     ▼
                          api.yaml loaded at runtime
                          routes wired dynamically
```

1. Define endpoints in `api.yaml` (method, path, handler name, auth, roles).
2. Run `gosupabase gen` to scaffold handler functions and server code.
3. Implement your business logic in the generated handler files.
4. The server loads `api.yaml` at startup and wires routes through a handler registry — no code regeneration needed when you add new endpoints.

## Next Steps

- [Install goSupaBase](/guide/installation)
- [Quick Start — build your first API in 60 seconds](/guide/quick-start)
