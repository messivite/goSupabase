# Configuration

goSupaBase resolves output directories with the following precedence (highest wins):

1. **CLI flags** — `--server-dir`, `--handlers-dir`
2. **`.gosupabase.yaml`** — created by `gosupabase setup`
3. **`api.yaml` output section**
4. **Defaults** — `server/`, `handlers/`

## `.gosupabase.yaml`

Created automatically by the setup wizard:

```yaml
output:
  serverDir: server
  handlersDir: handlers
deploy:
  provider: none
```

## `api.yaml` output section

```yaml
version: "1"
basePath: /api
output:
  serverDir: server
  handlersDir: handlers
```

## CLI flag overrides

```bash
gosupabase gen --server-dir pkg/server --handlers-dir pkg/handler
```

## Handlers-only mode

For projects with a custom server setup, generate only handler stubs and the registry:

```bash
gosupabase gen --handlers-only
```

This leaves your server code untouched.
