# Developer Flows

Copy-paste recipes for common scenarios.

## Flow 1: New API project from scratch

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

## Flow 2: Existing project integration

```bash
go mod init github.com/you/repo   # if not already
gosupabase init
go get github.com/messivite/gosupabase@latest && go mod tidy
gosupabase setup --from-file ./supabase.env
gosupabase add endpoint "GET /health"
gosupabase add endpoint "PATCH /tracks/:id" --auth
gosupabase gen --handlers-only
go test ./...
```

## Flow 3: Custom output layout

**Full** `gen` (writes `server.go` under the chosen dir) requires local **`middleware/`** and **`internal/yaml/`** — same as default `gen`. Library apps should use **`--handlers-only`** with custom handler dirs only:

```bash
gosupabase gen --handlers-only --handlers-dir pkg/handler
```

If you have a full scaffold and want custom paths for both:

```bash
gosupabase gen --server-dir pkg/server --handlers-dir pkg/handler
```

Output directories can also be set in `.gosupabase.yaml` or `api.yaml` — see [Configuration](/guide/configuration).

## Flow 4: Add an endpoint and re-generate

**After `gosupabase new` (or any repo with `middleware/` + `internal/yaml/`):**

```bash
gosupabase add endpoint "DELETE /tracks/:id" --auth
gosupabase gen
```

**Existing project using the library (`go get github.com/messivite/gosupabase`):**

```bash
gosupabase add endpoint "DELETE /tracks/:id" --auth
gosupabase gen --handlers-only
```

Only missing handler files are created; existing handlers are never overwritten. The server wires routes from `api.yaml` at runtime, so you can also add endpoints to YAML manually and register a handler — no `gen` required for routing changes.
