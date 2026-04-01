# goSupaBase

> Build Supabase-backed Go APIs fast with YAML-first endpoint design, runtime wiring, JWT auth, role guards, and hot-reload DX.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white&style=for-the-badge)](https://go.dev/)
[![Supabase](https://img.shields.io/badge/Supabase-Ready-3ECF8E?logo=supabase&logoColor=white&style=for-the-badge)](https://supabase.com/)
[![JWT](https://img.shields.io/badge/JWT-HS256%20%7C%20ES256-orange?style=for-the-badge)](#auth-middleware)
[![JWKS](https://img.shields.io/badge/JWKS-Auto%20Fetch-2563eb?style=for-the-badge)](#jwks-file-placement)
[![Hot Reload](https://img.shields.io/badge/Dev-gosupabase%20dev-16a34a?style=for-the-badge)](#quick-start)
[![CI](https://img.shields.io/github/actions/workflow/status/messivite/goSupabase/ci.yml?branch=main&style=for-the-badge&label=CI)](https://github.com/messivite/goSupabase/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/github/actions/workflow/status/messivite/goSupabase/ci.yml?branch=main&style=for-the-badge&label=Tests)](https://github.com/messivite/goSupabase/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/messivite/goSupabase/release.yml?style=for-the-badge&label=Release)](https://github.com/messivite/goSupabase/actions/workflows/release.yml)
[![Coverage](https://img.shields.io/codecov/c/github/messivite/goSupabase?style=for-the-badge&label=Coverage)](https://app.codecov.io/gh/messivite/goSupabase)

[![Router](https://img.shields.io/badge/Router-chi-f43f5e?style=flat-square)](https://github.com/go-chi/chi)
[![Codegen](https://img.shields.io/badge/Codegen-Handlers%20%2B%20Server-0ea5e9?style=flat-square)](#what-this-package-gives-you)
[![YAML First](https://img.shields.io/badge/API%20Design-YAML--First-7c3aed?style=flat-square)](#yaml-schema)
[![License](https://img.shields.io/badge/License-MIT-22c55e?style=flat-square)](LICENSE)
[![Last Commit](https://img.shields.io/github/last-commit/messivite/goSupabase?style=flat-square)](https://github.com/messivite/goSupabase/commits/main)

### Quick Links

- [Quick Start](#quick-start)
- [Setup](#setup)
- [Developer Flows](#developer-flows-copypaste)
- [Auth Middleware](#auth-middleware)
- [JWT Mode Changes](#jwt-mode-changes-what-happens)
- [YAML Schema](#yaml-schema)
- [Environment Variables](#environment-variables)
- [CI/CD and Releases](#cicd-and-releases)

## Features

- **YAML endpoint definitions** — single `api.yaml` as the source of truth
- **Code generation** — handler stubs and server wiring from YAML
- **Runtime routing** — `api.yaml` loaded at startup, no regeneration needed for new endpoints
- **Supabase JWT auth** — HS256/ES256 (JWKS) with claims context and role guards
- **Configurable output** — custom directories via flags, `.gosupabase.yaml`, or `api.yaml`
- **Handlers-only mode** — generate stubs without touching the server

## Quick Start

### New Project

```bash
go install github.com/messivite/goSupabase/cmd/gosupabase@latest

gosupabase new my-api
cd my-api
go mod tidy
gosupabase gen
go run ./cmd/server
```

### Existing Project

```bash
gosupabase init          # creates api.yaml + .env.example
gosupabase setup         # creates .env + .gosupabase.yaml
gosupabase add endpoint "POST /tracks" --auth
gosupabase gen
go run ./cmd/server
```

Hot-reload development:

```bash
gosupabase dev
```

If you see `zsh: command not found: gosupabase`, use one of these:

```bash
# Run directly from source (quickest)
go run ./cmd/gosupabase dev

# Build local binary in project root
go build -o gosupabase ./cmd/gosupabase
./gosupabase dev

# Install globally (requires Go bin in PATH)
go install ./cmd/gosupabase
gosupabase dev
```

If `go install` succeeds but command is still not found, add Go bin to zsh PATH:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

## Developer Flows (Copy/Paste)

### Flow 1: New API project from scratch

```bash
gosupabase new my-api
cd my-api
go mod tidy
gosupabase setup
gosupabase add endpoint "GET /users"
gosupabase add endpoint "POST /users" --auth
gosupabase gen
go build ./...
go run ./cmd/server
```

### Flow 2: Existing project integration

```bash
gosupabase init
gosupabase setup --from-file ./supabase.env
gosupabase add endpoint "GET /health"
gosupabase add endpoint "PATCH /tracks/:id" --auth
gosupabase gen --handlers-only
go test ./...
```

### Flow 3: Custom output layout

```bash
gosupabase gen --server-dir pkg/server --handlers-dir pkg/handler
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `gosupabase new <name>` | Scaffold a new project with all directories, `api.yaml`, `Makefile`, `go.mod` |
| `gosupabase init` | Initialize goSupabase in an existing project (creates `api.yaml`) |
| `gosupabase setup` | Interactive wizard for `.env` and `.gosupabase.yaml` |
| `gosupabase setup --from-file <path>` | Import config from an env-style file |
| `gosupabase add endpoint "METHOD /path" [--auth]` | Add an endpoint to `api.yaml` |
| `gosupabase gen [flags]` | Generate handler stubs and server code |
| `gosupabase dev` | Run `cmd/server` and auto-restart when watched files change |
| `gosupabase list` | List all defined endpoints |

### Gen Flags

```
--server-dir DIR       Override server output directory
--handlers-dir DIR     Override handlers output directory
--handlers-only        Generate only handler stubs (skip server)
```

## Setup

### Interactive Wizard (default)

Running `gosupabase setup` with no flags starts an interactive wizard that asks:

```
goSupabase interactive setup
----------------------------------------
Server port [8080]:
Supabase URL: https://abc.supabase.co
Supabase anon key: eyJ...
Supabase JWT secret: super-secret...
Include service role key? (server-side only, never expose publicly) [y/N]: n

Server output directory [server]:
Handlers output directory [handlers]:
```

The wizard writes two files:
- `.env` with your Supabase credentials
- `.gosupabase.yaml` with output directory preferences

**File conflict handling:** If `.env` or `.gosupabase.yaml` already exist, the wizard asks per file:
- **Overwrite** -- replace the file entirely with new values
- **Merge** -- add missing keys but keep existing values (`.env` only)
- **Skip** -- don't touch the file

### Import from File

For CI pipelines or teams that manage credentials externally:

```bash
gosupabase setup --from-file ./my-supabase.env
```

The file should be a standard `KEY=VALUE` format:

```env
PORT=8080
SUPABASE_URL=https://abc.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_JWT_SECRET=super-secret-jwt
SUPABASE_JWT_VALIDATION_MODE=auto
SERVER_DIR=pkg/server
HANDLERS_DIR=pkg/handler
```

The CLI validates required keys (`SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_JWT_SECRET`) and warns about any that are missing. Optional keys like `SERVER_DIR`, `HANDLERS_DIR`, and `SUPABASE_JWT_VALIDATION_MODE` map to runtime and output behavior.

The same conflict policy (overwrite/merge/skip) applies when target files already exist.

**Security note:** `SUPABASE_SERVICE_ROLE_KEY` is for server-side use only. Never expose it in client-side code or public endpoints. The setup wizard warns about this when the key is included.

## What This Package Gives You

- `api.yaml`-first API design (single source of truth for routes)
- Handler stub generation with automatic registry wiring
- Runtime route binding from YAML (no mandatory server regeneration for every new endpoint)
- Supabase JWT auth middleware (`HS256`) + role checks
- Output path precedence: flags > `.gosupabase.yaml` > `api.yaml` > defaults
- Two setup patterns:
  - Interactive wizard (`gosupabase setup`)
  - Guided import (`gosupabase setup --from-file <path>`)

## YAML Schema

```yaml
version: "1"
basePath: /api
output:
  serverDir: server
  handlersDir: handlers

endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false

  - method: POST
    path: /tracks
    handler: CreateTrack
    auth: true

  - method: DELETE
    path: /tracks/:id
    handler: DeleteTrack
    auth: true
    roles:
      - admin
```

### Endpoint Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `method` | string | yes | HTTP method (GET, POST, PATCH, PUT, DELETE) |
| `path` | string | yes | Route path (`:param` for path parameters) |
| `handler` | string | no | Handler function name (auto-derived if omitted) |
| `auth` | bool | no | Require Supabase JWT authentication |
| `roles` | []string | no | Allowed roles (e.g., `authenticated`, `admin`) |

### Handler Name Derivation

When `handler` is omitted, it's derived from method + path:

| Method | Path | Derived Name |
|--------|------|-------------|
| `POST` | `/tracks` | `PostTracks` |
| `GET` | `/tracks/:id` | `GetTracksById` |
| `DELETE` | `/tracks/:id` | `DeleteTracksById` |

## Runtime Routing

The server loads `api.yaml` at startup and wires routes dynamically:

1. Reads `api.yaml` and iterates over endpoints
2. Looks up each handler by name in the registry (`handlers.Get`)
3. Applies Supabase auth middleware for `auth: true` endpoints
4. Applies role guard middleware when `roles` is specified
5. Skips unregistered handlers with a warning log (no crash)

This means you can add endpoints to `api.yaml` and register handlers without regenerating server code.

### Handler Registration

Generated handlers register themselves via `init()`:

```go
package handlers

func init() {
    Register("CreateTrack", CreateTrack)
}

func CreateTrack(w http.ResponseWriter, r *http.Request) {
    // your implementation
}
```

## Auth Middleware

### How It Works

1. Extracts `Authorization: Bearer <token>` from the request
2. Verifies JWT based on `SUPABASE_JWT_VALIDATION_MODE`:
   - `auto` (default): detect from token `alg` (HS256 or ES256)
   - `jwks`: force JWKS/asymmetric validation only
   - `hs256`: force symmetric secret validation only
3. Validates token expiration
4. Injects claims into the request context

### JWT Mode Changes (What Happens)

You can switch validation behavior by changing `SUPABASE_JWT_VALIDATION_MODE` in `.env` and restarting the server.

| Mode | Behavior | Typical use |
|------|----------|-------------|
| `auto` | Uses token `alg`: `HS256` => secret, `ES256` => JWKS | Recommended default |
| `jwks` | Accepts only asymmetric (`ES256`) tokens via JWKS | Supabase modern projects |
| `hs256` | Accepts only symmetric (`HS256`) tokens via `SUPABASE_JWT_SECRET` | Legacy/self-managed JWT |

After changing mode:

```bash
go run ./cmd/server
```

If mode/token type mismatch occurs, auth endpoints return `401 invalid token`.

### JWKS File Placement

You do not need to store `jwks.json` in your project.

- Runtime fetches keys from:
  - `SUPABASE_URL/auth/v1/.well-known/jwks.json`
- Local `jwks.json` is only optional for manual debugging.

### Accessing Claims

```go
import "github.com/messivite/goSupabase/auth"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    claims := auth.GetClaims(r.Context())
    if claims != nil {
        userID := claims.Subject
        role   := claims.Role
        email  := claims.Email
    }
}
```

### Claims Fields

| Field | JSON | Description |
|-------|------|-------------|
| `Subject` | `sub` | User ID |
| `Role` | `role` | Supabase role (e.g., `authenticated`, `admin`) |
| `Email` | `email` | User email |
| `Audience` | `aud` | Token audience |
| `ExpiresAt` | `exp` | Expiration timestamp |

### Role Guard

Endpoints with `roles` in `api.yaml` are protected by the role guard middleware. Only requests with a matching role in the JWT claims are allowed; others receive a `403 Forbidden`.

## Custom Output Paths

Output directory resolution follows this precedence (highest to lowest):

1. **CLI flags**: `--server-dir`, `--handlers-dir`
2. **`.gosupabase.yaml`**:
   ```yaml
   output:
     serverDir: pkg/server
     handlersDir: pkg/handler
   ```
3. **`api.yaml` output section**:
   ```yaml
   output:
     serverDir: server
     handlersDir: handlers
   ```
4. **Defaults**: `server/`, `handlers/`

## Handlers-Only Mode

For projects with a custom server setup:

```bash
gosupabase gen --handlers-only
```

This generates only handler stubs and the registry, leaving your server code untouched.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | no | Server port (default: `8080`) |
| `SUPABASE_URL` | yes | Your Supabase project URL |
| `SUPABASE_ANON_KEY` | yes | Supabase anonymous/public key |
| `SUPABASE_SERVICE_ROLE_KEY` | no | Service role key (server-side only, never expose) |
| `SUPABASE_JWT_SECRET` | yes | JWT secret for token verification |
| `SUPABASE_JWT_VALIDATION_MODE` | no | `auto` (default), `jwks`, or `hs256` |

Copy `.env.example` to `.env` and fill in your credentials.

## Music API Examples

The default `api.yaml` includes sample endpoints for a music API:

| Method | Path | Handler | Auth | Roles |
|--------|------|---------|------|-------|
| GET | /health | Health | no | - |
| GET | /tracks | ListTracks | no | - |
| POST | /tracks | CreateTrack | yes | - |
| PATCH | /tracks/:id | UpdateTrack | yes | - |
| DELETE | /tracks/:id | DeleteTrack | yes | admin |
| GET | /playlists | ListPlaylists | no | - |
| POST | /playlists | CreatePlaylist | yes | - |

## Project Structure

```
├── cmd/
│   ├── gosupabase/main.go    # CLI entry point
│   └── server/main.go        # Server entry point
├── internal/
│   ├── yaml/api.go            # YAML schema parsing
│   └── scaffold/
│       ├── generator.go       # Code generation engine
│       └── templates.go       # Go templates for codegen
├── middleware/supabase.go     # JWT auth + role guard
├── auth/claims.go             # Claims model + context helpers
├── handlers/registry.go       # Handler registration map
├── server/server.go           # Runtime route wiring
├── config/env.go              # Env config + output path resolution
├── api/index.go               # Serverless entry (Vercel-compatible)
├── api.yaml                   # Endpoint definitions
├── .env.example               # Environment template
└── Makefile
```

## Development

```bash
go mod tidy
go build ./...
go test ./...
```

## CI/CD and Releases

This repository includes two GitHub Actions workflows:

- `CI` (`.github/workflows/ci.yml`)
  - Runs on push/PR to `main`
  - Checks `go mod tidy` consistency
  - Runs `go build ./...`, `go test ./...`, and `go vet ./...`

- `Release` (`.github/workflows/release.yml`)
  - **Automatic release** when pushing tags like `v1.2.3`
  - **Manual release** from GitHub Actions UI via `workflow_dispatch`
  - Includes a required **quality gate** (`build + test + vet + tidy check`)
  - Release artifacts are built only if the quality gate succeeds
  - Builds binaries for Linux/macOS/Windows and uploads `checksums.txt`

### Automatic release (tag-based)

```bash
git tag v0.1.0
git push origin v0.1.0
```

### Manual release (GitHub UI)

1. Go to **Actions** -> **Release** workflow
2. Click **Run workflow**
3. Provide:
   - `tag` (optional, example: `v0.1.0`)
   - `bump` (`patch`/`minor`/`major`) when `tag` is empty
   - `target` (`main` by default)
   - `make_latest` (`true/false`)

If `tag` is empty, workflow auto-calculates the next semver from the latest `v*` tag:
- `patch` -> `vX.Y.(Z+1)`
- `minor` -> `vX.(Y+1).0`
- `major` -> `v(X+1).0.0`

Then it creates the tag, runs quality gate, builds artifacts, and publishes a GitHub Release.

## License

MIT
