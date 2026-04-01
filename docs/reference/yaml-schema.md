# YAML Schema

All endpoints are defined in `api.yaml`, the single source of truth for your API routes.

## Full Example

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

## Endpoint Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `method` | string | yes | HTTP method (`GET`, `POST`, `PATCH`, `PUT`, `DELETE`) |
| `path` | string | yes | Route path (`:param` for path parameters) |
| `handler` | string | no | Handler function name (auto-derived if omitted) |
| `auth` | bool | no | Require Supabase JWT authentication |
| `roles` | `[]string` | no | Allowed roles (e.g., `authenticated`, `admin`) |

## Handler Name Derivation

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

## Handler Registration

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
