# Auth & JWT

goSupaBase includes middleware for Supabase JWT authentication with HS256 and ES256 (JWKS) support.

## How It Works

1. Extracts `Authorization: Bearer <token>` from the request
2. Verifies JWT based on `SUPABASE_JWT_VALIDATION_MODE`:
   - `auto` (default): detect from token `alg` header (HS256 or ES256)
   - `jwks`: force JWKS/asymmetric validation only
   - `hs256`: force symmetric secret validation only
3. Validates token expiration
4. Injects claims into the request context

## JWT Validation Modes

| Mode | Behavior | Typical use |
|------|----------|-------------|
| `auto` | Reads token `alg` — `HS256` uses secret, `ES256` uses JWKS | Recommended default |
| `jwks` | Accepts only asymmetric (`ES256`) tokens via JWKS | Supabase modern projects |
| `hs256` | Accepts only symmetric (`HS256`) tokens via `SUPABASE_JWT_SECRET` | Legacy or self-managed JWT |

Switch modes by changing `SUPABASE_JWT_VALIDATION_MODE` in `.env` and restarting:

```bash
go run ./cmd/server
```

If mode and token type don't match, auth endpoints return `401 invalid token`.

## JWKS Key Fetching

You do **not** need to store `jwks.json` in your project.

- At runtime, keys are fetched from `SUPABASE_URL/auth/v1/.well-known/jwks.json`
- A local `jwks.json` file is only for optional manual debugging

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
