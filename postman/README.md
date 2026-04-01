# Postman collection

`goSupabase.postman_collection.json` is a Postman Collection v2.1 export you can import into Postman, Insomnia, or similar tools.

## Import

1. Postman → **Import** → **File** → select `goSupabase.postman_collection.json`.
2. Open the collection **Variables** tab (or edit collection variables):
   - `baseUrl` — API root (default `http://localhost:8080`).
   - `accessToken` — Supabase **user** JWT (`session.access_token`), not the anon key.
   - `trackId` — path parameter for PATCH/DELETE `/api/tracks/:id`.

## Sharing

- **In-repo JSON** (this file): versioned with Git; good for open source and reviews.
- **Postman workspace / link**: optional; useful for teams, but keep the JSON export as the source of truth if you want the same behavior as `api.yaml`.

Replace or extend this file with your own export if you already have a richer collection.
