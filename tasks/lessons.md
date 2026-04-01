# goSupaBase – Lessons Learned

_Updated after each correction or insight during development._

## 2026-04-01

- Supabase auth tokens may use `ES256`/`RS256` (asymmetric) instead of only `HS256`.
- Rule: never assume JWT algorithm; always branch verify flow by `alg` from token header.
- Rule: when `alg` is asymmetric, verify with JWKS using `kid` from token header.
