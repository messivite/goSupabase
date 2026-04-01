# CI/CD & Releases

goSupaBase includes two GitHub Actions workflows for continuous integration and automated releases.

## CI Workflow

**File:** `.github/workflows/ci.yml`

Runs on every push and pull request to `main`:

- Verifies `go mod tidy` consistency
- `go build ./...`
- `go test ./... -count=1`
- `go vet ./...`
- Coverage upload to Codecov

Tests run on a matrix of Go 1.22 / 1.23 across Ubuntu and macOS.

## Release Workflow

**File:** `.github/workflows/release.yml`

### Automatic release (tag-based)

```bash
git tag v0.1.0
git push origin v0.1.0
```

Pushing a `v*` tag triggers the release pipeline automatically.

### Manual release (GitHub UI)

1. Go to **Actions** → **Release** workflow
2. Click **Run workflow**
3. Provide:
   - `tag` — optional (example: `v0.1.0`). Leave empty for auto-bump.
   - `bump` — `patch` / `minor` / `major` (used when `tag` is empty)
   - `target` — branch or commit (default: `main`)
   - `make_latest` — `true` / `false`

When `tag` is empty, the workflow auto-calculates the next semver:

| Bump | Formula |
|------|---------|
| `patch` | `vX.Y.(Z+1)` |
| `minor` | `vX.(Y+1).0` |
| `major` | `v(X+1).0.0` |

### Release pipeline

1. **Create tag** — resolves or auto-bumps the version
2. **Quality gate** — build + test + vet + tidy check
3. **Build artifacts** — cross-compile for Linux/macOS/Windows + `checksums.txt`
4. **Publish** — creates a GitHub Release with all binaries

## Documentation Deployment

Documentation is built and deployed to GitHub Pages automatically on every push to `main` (when files under `docs/` change). See `.github/workflows/docs.yml`.

You can also build docs locally:

```bash
npm ci
npm run docs:build
```
