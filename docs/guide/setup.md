# Setup Wizard

The `gosupabase setup` command configures your local environment. It has two modes: interactive (default) and file-based import.

## Interactive Wizard

```bash
gosupabase setup
```

The wizard prompts for:

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

Add deploy scaffolding (vercel.json, fly.toml, …)? Optional — skip if you will configure hosting later [y/N]:
```

Deploy is **optional**. Press **Enter** (or answer `n`) to skip — no deploy questions follow. If you answer `y`, the wizard asks for a target and, for Fly, an optional app name. See [Deployment](/advanced/deployment) for details.

### What it writes

- `.env` with your Supabase credentials
- `.gosupabase.yaml` with output paths and `deploy.provider`
- Provider-specific deploy files (if you enabled deploy scaffolding)

### File conflict handling

If `.env` or `.gosupabase.yaml` already exist, the wizard asks per file:

| Choice | Behavior |
|--------|----------|
| **Overwrite** | Replace the file entirely with new values |
| **Merge** | Add missing keys but keep existing values (`.env` only) |
| **Skip** | Don't touch the file |

## Import from File

For CI pipelines or teams that manage credentials externally:

```bash
gosupabase setup --from-file ./my-supabase.env
```

The file uses standard `KEY=VALUE` format:

```ini
PORT=8080
SUPABASE_URL=https://abc.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_JWT_SECRET=super-secret-jwt
SUPABASE_JWT_VALIDATION_MODE=auto
SERVER_DIR=pkg/server
HANDLERS_DIR=pkg/handler

# optional deploy keys
DEPLOY_TARGET=vercel
FLY_APP_NAME=
DEPLOY_OVERWRITE=false
```

Required keys: `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_JWT_SECRET`. The CLI warns about any that are missing but continues with available values.

Deploy-related keys (`DEPLOY_TARGET`, `FLY_APP_NAME`, `DEPLOY_OVERWRITE`) are documented under [Deployment](/advanced/deployment).

::: warning Security
`SUPABASE_SERVICE_ROLE_KEY` is for server-side use only. Never expose it in client-side code or public endpoints.
:::
