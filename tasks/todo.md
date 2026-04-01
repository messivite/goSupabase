# goSupaBase – Task Tracker

## Phase 1: MVP

- [x] Module skeleton and directory layout
- [x] api.yaml schema parsing, validation, endpoint CRUD
- [x] Config precedence resolution (flags > .gosupabase.yaml > api.yaml > defaults)
- [x] Auth middleware (Supabase JWT HS256) + claims context + role guard
- [x] Runtime router (chi, api.yaml wiring, registry lookup)
- [x] Codegen engine (handlers, server, handlers-only mode)
- [x] CLI commands (new, init, setup, add, gen, list)
- [x] Unit tests (yaml, config, middleware)
- [x] README, .env.example, sample api.yaml, CHANGELOG
- [x] Verification: build, test, smoke

## Review

- `go build ./...` -- PASS
- `go test ./...` -- 20/20 tests PASS (yaml: 9, config: 6, middleware: 10)
- `go vet ./...` -- PASS
- CLI smoke: `list`, `new`, `add endpoint`, duplicate detection, `gen --handlers-only` -- all PASS
- Generated handlers compile cleanly after codegen
- Config precedence verified: flags > .gosupabase.yaml > api.yaml > defaults
