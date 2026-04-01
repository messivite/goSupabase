# Deployment

Everything here is **optional** during `gosupabase setup`. You can skip deploy scaffolding entirely — the interactive default is **no**, and `--from-file` simply omits `DEPLOY_TARGET`.

## Supported Targets

| Target | Generated files | What it does |
|--------|----------------|--------------|
| `vercel` | `vercel.json`, `DEPLOY.md` | Serverless entry via `api/index.go`; set env vars in Vercel |
| `fly` | `fly.toml`, `DEPLOY.md` | Container on Fly.io; optional app name placeholder; `fly secrets` + `fly deploy` |
| `railway` | `railway.toml`, `DEPLOY.md` | Tune start command and variables in Railway |
| `render` | `render.yaml`, `DEPLOY.md` | Blueprint stub; adjust repo URL and create a Web Service |
| `none` | — | No extra files; `.gosupabase.yaml` records `deploy.provider: none` |

`DEPLOY.md` is a short checklist for the chosen provider. **Never** commit real secrets — configure them in each platform's UI.

## Interactive Wizard

1. After server/handler directories, the wizard asks: `Add deploy scaffolding … [y/N]`.
   - **Enter** or `n` → skip. No deploy files, `deploy.provider: none`.
2. If `y`: choose `vercel` / `fly` / `railway` / `render` / `none` (empty defaults to `none`).
3. For **Fly**, you can set an app name or leave it empty (placeholder `my-gosupabase-app`).

## `setup --from-file`

| Key | Meaning |
|-----|---------|
| `DEPLOY_TARGET` | `vercel`, `fly`, `railway`, `render`, or empty / `none` |
| `FLY_APP_NAME` | Optional; written into `fly.toml` when target is `fly` |
| `DEPLOY_OVERWRITE` | `true` / `false` — replace existing deploy files instead of skipping |

## Provider Notes

### Vercel

Ensure `api/index.go` exists (created by `gosupabase new` or `gen`). Link the repo in Vercel, set `SUPABASE_*` and `PORT` environment variables, and deploy.

### Fly.io

Follow the generated `DEPLOY.md`:

```bash
fly launch
fly secrets set SUPABASE_URL=... SUPABASE_ANON_KEY=... SUPABASE_JWT_SECRET=...
fly deploy
```

### Railway

Point Railway at this repo. Align the `railway.toml` start command with your `cmd/server` binary build. Set environment variables in the Railway dashboard.

### Render

Use `render.yaml` as a starting blueprint. Connect the Git repo and set environment variables in the Render dashboard.

## Config and Conflicts

- `.gosupabase.yaml` includes `deploy.provider` so `gosupabase gen` and future tooling stay aligned with your host choice.
- If deploy files already exist and you run setup again, unchanged files are **skipped** unless you choose overwrite or set `DEPLOY_OVERWRITE=true` in `--from-file` mode.
