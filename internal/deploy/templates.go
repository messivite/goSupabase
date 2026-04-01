package deploy

import "fmt"

const vercelJSON = `{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "rewrites": [
    { "source": "/(.*)", "destination": "/api" }
  ],
  "functions": {
    "api/**/*.go": {
      "maxDuration": 30
    }
  }
}
`

func flyTOML(appName string) string {
	return fmt.Sprintf(`# Rename app after: fly apps create <name>
app = "%s"
primary_region = "iad"

[build]

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = "stop"
  auto_start_machines = true
  min_machines_running = 0
  processes = ["app"]

[[vm]]
  memory = "256mb"
  cpu_kind = "shared"
  cpus = 1
`, appName)
}

const railwayTOML = `# Railway picks up Go via NIXPACKS. Set env vars in the Railway dashboard.
[build]
builder = "NIXPACKS"

[deploy]
startCommand = "./server"
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
`

const renderYAML = `# Render Blueprint — adjust name and branch as needed
services:
  - type: web
    name: gosupabase-api
    runtime: go
    repo: https://github.com/your-org/your-repo
    branch: main
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: PORT
        value: "8080"
      - key: GO_VERSION
        value: "1.22"
`

func deployMarkdown(provider string) string {
	common := "# Deployment — goSupabase\n\n" +
		"This project ships with a standard Go server (`cmd/server`) and an optional Vercel-style handler (`api/index.go`).\n\n" +
		"Set the same environment variables everywhere (see `.env.example`).\n\n"
	switch provider {
	case ProviderVercel:
		return common + "## Vercel\n\n" +
			"1. Push this repo to GitHub and import it in the [Vercel dashboard](https://vercel.com).\n" +
			"2. **Root directory**: repository root (where `go.mod` lives).\n" +
			"3. **Environment variables**: add `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_JWT_SECRET`, `SUPABASE_JWT_VALIDATION_MODE`, and `PORT` if needed.\n" +
			"4. The serverless entry is `api/index.go` (`Handler`). `vercel.json` rewrites all routes to `/api`.\n\n"
	case ProviderFly:
		return common + "## Fly.io\n\n" +
			"1. Install the [Fly CLI](https://fly.io/docs/hands-on/install-flyctl/).\n" +
			"2. Edit `fly.toml` — set a unique `app` name (or run `fly launch` and merge settings).\n" +
			"3. Set secrets: `fly secrets set SUPABASE_URL=... SUPABASE_JWT_SECRET=...` (and other keys).\n" +
			"4. Deploy: `fly deploy` (add a Dockerfile or align build with Fly; process should run `cmd/server`).\n\n"
	case ProviderRailway:
		return common + "## Railway\n\n" +
			"1. Create a project and deploy from GitHub.\n" +
			"2. Set **Start command** to your compiled binary or `go run ./cmd/server` (production: prefer a built binary).\n" +
			"3. Add all Supabase-related env vars in the Railway **Variables** tab.\n\n"
	case ProviderRender:
		return common + "## Render\n\n" +
			"1. Edit `render.yaml` with your repo URL and service name.\n" +
			"2. Create a **Blueprint** from this file in the Render dashboard, or a **Web Service** with build `go build -o server ./cmd/server` and start `./server`.\n" +
			"3. Add environment variables in the Render dashboard.\n\n"
	default:
		return common
	}
}
