# Installation

## Prerequisites

- **Go 1.22+** — [download](https://go.dev/dl/)
- A Supabase project (for JWT auth) — [supabase.com](https://supabase.com/)

## Install the CLI

```bash
go install github.com/messivite/gosupabase/cmd/gosupabase@latest
```

Verify:

```bash
gosupabase help
```

## Troubleshooting

### `zsh: command not found: gosupabase`

The `go install` binary lands in `$(go env GOPATH)/bin`. If that directory is not in your `PATH`, the shell won't find it. Quick fixes:

```bash
# Option 1: run from source (no install needed)
go run ./cmd/gosupabase dev

# Option 2: build a local binary
go build -o gosupabase ./cmd/gosupabase
./gosupabase dev

# Option 3: add Go bin to PATH permanently
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

After option 3, `gosupabase` works globally in any terminal session.
