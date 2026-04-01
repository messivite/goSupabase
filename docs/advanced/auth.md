# Auth & JWT

goSupaBase includes middleware for Supabase JWT authentication with **HS256** (symmetric) and **ES256** (asymmetric, JWKS) support. This matches how Supabase issues tokens: see the official guides on [JWTs](https://supabase.com/docs/guides/auth/jwts) and [JWT signing keys](https://supabase.com/docs/guides/auth/signing-keys) (legacy JWT secret vs asymmetric signing keys).

## How do I verify Supabase JWTs and JWKS?

1. Put **`SUPABASE_URL`**, **`SUPABASE_JWT_SECRET`**, and **`SUPABASE_ANON_KEY`** in `.env` (see [Environment variables](/reference/environment-variables)). Set **`SUPABASE_JWT_VALIDATION_MODE=auto`** unless you know you need `jwks` or `hs256` only.
2. Call your API with **`Authorization: Bearer`** plus the **user access token** (JWT string from Supabase Auth). Do not send the anon key as the Bearer token when you want to identify the signed-in user.
3. The middleware reads the JWT **header**: if **`alg`** is **`HS256`**, it verifies with **`SUPABASE_JWT_SECRET`**; if **`ES256`**, it loads **`{SUPABASE_URL}/auth/v1/.well-known/jwks.json`**, matches **`kid`** to a **P-256** key, and verifies the signature. No committed `jwks.json` file is required.

Details, troubleshooting, and cache behavior are in the sections below — start with [Token algorithms (Supabase alignment)](#token-algorithms-supabase-alignment) and [JWKS endpoint and key shape](#jwks-endpoint-and-key-shape).

## How It Works

1. Extracts `Authorization: Bearer <token>` from the request
2. Reads the JWT **header** (first segment, base64url) for `alg` and `kid`
3. Verifies the signature using either `SUPABASE_JWT_SECRET` (HS256) or keys from Supabase’s JWKS (ES256), according to `SUPABASE_JWT_VALIDATION_MODE`
4. Validates token expiration (`exp`)
5. Injects claims into the request context

## Token algorithms (Supabase alignment)

Supabase can sign access tokens with:

| `alg` in JWT header | Verification in goSupaBase | What you need in `.env` |
|---------------------|----------------------------|---------------------------|
| **HS256** | HMAC with `SUPABASE_JWT_SECRET` | Correct JWT secret from Dashboard → Project Settings → API (JWT secret) |
| **ES256** | ECDSA P-256 public key from JWKS, matched by `kid` | `SUPABASE_URL` must reach your project; JWKS is fetched automatically |

The **anon** and **service role** keys are not used to verify signatures — they identify the client. Verification always uses either the shared JWT secret (HS256) or the public keys published in JWKS (ES256), as described in Supabase’s docs.

::: tip Default
Use **`SUPABASE_JWT_VALIDATION_MODE=auto`** unless you have a reason to force one path. It picks the verifier based on the token’s `alg` header, which is the least surprising behavior when Supabase rotates or mixes signing setups.
:::

## JWT Validation Modes

| Mode | Behavior | Typical use |
|------|----------|-------------|
| `auto` | Uses token `alg`: `HS256` → secret, `ES256` → JWKS | **Recommended** — matches token type automatically |
| `jwks` | Only **ES256** tokens; HS256 is rejected | Force asymmetric-only (e.g. after migrating off legacy JWT secret) |
| `hs256` | Only **HS256** tokens; ES256 is rejected | Legacy / symmetric-only environments |

Switch modes by changing `SUPABASE_JWT_VALIDATION_MODE` in `.env` and restarting:

```bash
go run ./cmd/server
```

If the mode does not match the token’s algorithm, protected routes return **`401`** with `invalid token`.

### Common mismatches

| Symptom | Likely cause |
|---------|----------------|
| `401 invalid token` with a valid browser/session token | Mode is `hs256` but Supabase now issues **ES256** (or vice versa) — switch to `auto` or the correct mode |
| `401` only for some users / after key rotation | **`kid`** in the JWT header does not match any key yet returned from JWKS — wait for cache expiry (see below) or confirm URL |
| Works locally, fails in production | Wrong **`SUPABASE_URL`** or secret for that environment |

## JWKS endpoint and key shape

You do **not** need to commit `jwks.json` in your repo for production verification.

- **URL used at runtime:** `{SUPABASE_URL}/auth/v1/.well-known/jwks.json`  
  (Same document Supabase documents for discovering signing keys; see [signing keys](https://supabase.com/docs/guides/auth/signing-keys).)

The middleware decodes the standard **JWK Set** JSON: a top-level `keys` array. For **ES256**, goSupaBase only loads keys that match all of:

- `kty` = `"EC"`
- `crv` = `"P-256"`
- Non-empty `kid`, `x`, and `y` (base64url-encoded coordinates)

Other key types (e.g. RSA `RS256`) are **not** used by this middleware today — tokens must be **ES256** with **P-256** if you rely on JWKS verification.

The JWT header’s **`kid`** must match a `kid` in that set so the correct public key is selected (important when Supabase rotates keys).

### Caching

JWKS responses are cached **per `SUPABASE_URL`** for about **5 minutes** to avoid hammering the endpoint. After rotation, allow a short window or restart the process if you need immediate pickup.

### Local `jwks.json`

A file named `jwks.json` in your project is **not** read by the server automatically. It is only useful for manual inspection or debugging (e.g. comparing `kid` and key material with what the live endpoint returns).

## How this maps to Supabase Dashboard

1. **Project URL** → `SUPABASE_URL` (must be the project ref URL, e.g. `https://xxxx.supabase.co`).
2. **JWT secret** (legacy symmetric) → `SUPABASE_JWT_SECRET` for HS256 verification.
3. **Asymmetric signing keys** → no manual JWK paste required; the middleware fetches `/.well-known/jwks.json` and verifies ES256 tokens.

If you recently changed signing strategy on Supabase’s side, re-read their [JWT signing keys](https://supabase.com/docs/guides/auth/signing-keys) page and prefer **`auto`** until everything stabilizes.

## Accessing Claims

```go
import "github.com/messivite/gosupabase/auth"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    claims := auth.GetClaims(r.Context())
    if claims != nil {
        userID := claims.Subject
        role   := claims.Role
        email  := claims.Email
    }
}
```

### Claims Fields

| Field | JSON key | Description |
|-------|----------|-------------|
| `Subject` | `sub` | User ID |
| `Role` | `role` | Supabase role (e.g., `authenticated`, `admin`) |
| `Email` | `email` | User email |
| `Audience` | `aud` | Token audience |
| `ExpiresAt` | `exp` | Expiration timestamp |

## Role Guard

Endpoints with `roles` in `api.yaml` are protected by the role guard middleware. Only requests with a matching role in the JWT claims are allowed; others receive `403 Forbidden`.

```yaml
endpoints:
  - method: DELETE
    path: /tracks/:id
    handler: DeleteTrack
    auth: true
    roles:
      - admin
```
