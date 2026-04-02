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

Full `gen` still requires **`middleware/`** and **`internal/yaml/`** under your module root; otherwise the command fails and suggests `--handlers-only`.

## Handlers-only mode

For apps that depend on **`github.com/messivite/gosupabase`** and run the published server from `cmd/server`, generate only handler stubs and the registry:

```bash
gosupabase gen --handlers-only
```

No local `server.go` is written, so you do not need vendored `middleware` / `internal/yaml` in your module. You can combine with `--handlers-dir` (and optionally `--server-dir` for future full-gen use) as needed.
