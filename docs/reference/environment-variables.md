# Environment Variables

## Required Variables

| Variable | Description |
|----------|-------------|
| `SUPABASE_URL` | Your Supabase project URL |
| `SUPABASE_ANON_KEY` | Supabase anonymous/public key |
| `SUPABASE_JWT_SECRET` | JWT secret for token verification |

## Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `SUPABASE_SERVICE_ROLE_KEY` | — | Service role key (server-side only, **never expose**) |
| `SUPABASE_JWT_VALIDATION_MODE` | `auto` | `auto`, `jwks`, or `hs256` — see [Auth & JWT](/advanced/auth) for how this interacts with HS256 vs ES256 tokens and JWKS |

`SUPABASE_URL` must be the project API URL (e.g. `https://xxxx.supabase.co`) so JWKS can be loaded from `/auth/v1/.well-known/jwks.json` when tokens use **ES256**.

## Setup

Copy `.env.example` to `.env` and fill in your credentials, or use the setup wizard:

```bash
gosupabase setup
```

See [Setup Wizard](/guide/setup) for interactive and file-based options.

::: warning Security
`SUPABASE_SERVICE_ROLE_KEY` is for server-side use only. Never expose it in client-side code, public endpoints, or version control.
:::
