# Changelog

## [Unreleased]

### Added
- Initial project scaffold
- YAML-based endpoint definition and parsing
- Supabase JWT (HS256) auth middleware with role guard
- Code generation for handlers and server wiring
- Runtime api.yaml route loading via chi
- CLI commands: new, init, setup, add, gen, list
- Config precedence: flags > .gosupabase.yaml > api.yaml output > defaults
- Music API sample endpoints (tracks, playlists)
