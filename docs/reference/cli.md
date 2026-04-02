# CLI Commands

## Overview

| Command | Description |
|---------|-------------|
| `gosupabase new <name>` | Scaffold a new project with all directories, `api.yaml`, `Makefile`, `go.mod` |
| `gosupabase init` | Initialize in an existing project: `api.yaml`, `.env.example`, and `cmd/server/main.go` if `go.mod` exists |
| `gosupabase setup` | Interactive wizard for `.env`, `.gosupabase.yaml`, and optional deploy templates |
| `gosupabase setup --from-file <path>` | Import config from an env-style file |
| `gosupabase add endpoint "METHOD /path" [--auth]` | Add an endpoint to `api.yaml` |
| `gosupabase gen [flags]` | Generate handler stubs and server code |
| `gosupabase dev` | Run `cmd/server` and auto-restart on file changes |
| `gosupabase list` | List all defined endpoints |
| `gosupabase help` | Show usage information |

## `gosupabase new`

```bash
gosupabase new my-api
```

Creates a directory with:
- `go.mod` (module `github.com/example/my-api`)
- `api.yaml` with sample endpoints
- `cmd/server/main.go` entry point
- `Makefile`
- `.env.example`

## `gosupabase init`

```bash
go mod init github.com/you/your-api   # required first
gosupabase init
```

Creates `api.yaml` (with a health endpoint), `.env.example`, and **`cmd/server/main.go`** when `go.mod` is present. The entrypoint imports `github.com/messivite/gosupabase/server` and `config` and blank-imports **your** `handlers` package — run `go get github.com/messivite/gosupabase@latest && go mod tidy` next.

If there is no `go.mod`, `init` prints a hint to create one and run `init` again.

## `gosupabase setup`

```bash
gosupabase setup                        # interactive wizard
gosupabase setup --from-file config.env # import from file
```

See [Setup Wizard](/guide/setup) for full details.

## `gosupabase add`

```bash
gosupabase add endpoint "GET /tracks"
gosupabase add endpoint "POST /tracks" --auth
gosupabase add endpoint "DELETE /tracks/:id" --auth
```

Appends the endpoint to `api.yaml`. Handler names are auto-derived from method + path when not specified.

## `gosupabase gen`

```bash
gosupabase gen
gosupabase gen --handlers-only
gosupabase gen --server-dir pkg/server --handlers-dir pkg/handler
```

### Flags

| Flag | Description |
|------|-------------|
| `--server-dir DIR` | Override server output directory |
| `--handlers-dir DIR` | Override handlers output directory |
| `--handlers-only` | Generate only handler stubs (skip server) |

Existing handler files are never overwritten — only missing ones are created.

## `gosupabase dev`

```bash
gosupabase dev
```

Builds and runs `cmd/server`, then watches for changes in `.go` files and `api.yaml`. Restarts the server automatically on change.

## `gosupabase list`

```bash
gosupabase list
```

Prints a table of all endpoints defined in `api.yaml`:

```
METHOD  PATH              HANDLER          AUTH  ROLES
GET     /health           Health           no    -
POST    /tracks           CreateTrack      yes   -
DELETE  /tracks/:id       DeleteTrack      yes   admin
```
