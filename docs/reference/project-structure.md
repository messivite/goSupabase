# Project Structure

A typical goSupaBase project looks like this:

```
├── cmd/
│   ├── gosupabase/main.go    # CLI entry point
│   └── server/main.go        # Server entry point
├── internal/
│   ├── yaml/api.go            # YAML schema parsing
│   ├── deploy/                # Deploy config generation
│   └── scaffold/
│       ├── generator.go       # Code generation engine
│       └── templates.go       # Go templates for codegen
├── middleware/supabase.go     # JWT auth + role guard
├── auth/claims.go             # Claims model + context helpers
├── handlers/
│   ├── registry.go            # Handler registration map
│   ├── health.go              # Generated handler stub
│   └── create_track.go        # Generated handler stub
├── server/server.go           # Runtime route wiring
├── config/env.go              # Env config + output path resolution
├── api/index.go               # Serverless entry (Vercel-compatible)
├── api.yaml                   # Endpoint definitions
├── .env                       # Local environment (git-ignored)
├── .env.example               # Environment template
├── .gosupabase.yaml           # Output paths + deploy provider
└── Makefile
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/gosupabase/` | CLI binary — `new`, `init`, `setup`, `add`, `gen`, `dev`, `list` |
| `cmd/server/` | Server binary — loads `.env`, starts the HTTP server |
| `handlers/` | Generated handler stubs + registry (safe to edit; never overwritten) |
| `server/` | Generated server with chi router + runtime YAML wiring |
| `middleware/` | Supabase JWT verification and role guard |
| `auth/` | JWT claims model and request-context helpers |
| `config/` | Environment loading and output-path resolution |
| `internal/scaffold/` | Code generation engine and Go templates |
| `internal/yaml/` | `api.yaml` parser |
| `internal/deploy/` | Deploy config file generation (Vercel, Fly, Railway, Render) |
| `api/` | Vercel-compatible serverless entry point |
