---
layout: home

hero:
  name: goSupaBase
  text: Build Go APIs fast with Supabase
  tagline: YAML-first endpoint design, runtime wiring, JWT auth, role guards, and hot-reload DX.
  image:
    src: https://cdn.simpleicons.org/supabase/3ECF8E
    alt: goSupaBase
  actions:
    - theme: brand
      text: Get Started
      link: /guide/introduction
    - theme: alt
      text: View on GitHub
      link: https://github.com/messivite/goSupabase

features:
  - icon: 📝
    title: YAML-First API Design
    details: Define endpoints in a single api.yaml — the source of truth for routes, auth, and role guards.
  - icon: ⚡
    title: Code Generation
    details: Generate handler stubs from YAML; full local server wiring only when your module includes middleware and yaml packages — library apps use handlers-only mode.
  - icon: 🔄
    title: Runtime Routing
    details: api.yaml is loaded at startup. Add endpoints without regenerating server code.
  - icon: 🔐
    title: Supabase JWT Auth
    details: HS256 and ES256 (JWKS) token verification with claims context and role-based guards.
  - icon: 🔥
    title: Hot-Reload Dev Server
    details: Run gosupabase dev — auto-restarts when Go files or api.yaml change.
  - icon: 🚀
    title: Deploy Anywhere
    details: Scaffold Vercel, Fly.io, Railway, or Render configs in seconds.
---
