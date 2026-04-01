# Postman collection

A starter **Postman Collection v2.1** lives in the repo at [`postman/goSupabase.postman_collection.json`](https://github.com/messivite/goSupabase/blob/main/postman/goSupabase.postman_collection.json).

## Import

1. Postman → **Import** → **File** → choose `postman/goSupabase.postman_collection.json`.
2. Set collection variables:
   - **`baseUrl`** — e.g. `http://localhost:8080`
   - **`accessToken`** — Supabase user JWT from Auth (`session.access_token`)
   - **`trackId`** — for `/api/tracks/:id` routes

Routes follow `api.yaml`: public `GET`s under `/api/...`, and `POST`/`PATCH`/`DELETE` with `Authorization: Bearer …` where `auth: true`.

## Export and share

- **Commit the JSON** in the repo (as here) so it stays in sync with `api.yaml` and is easy to review in PRs.
- You can also share a **Postman workspace link** for your team; keeping the exported file in Git is still recommended as the portable, versioned copy.

See [`postman/README.md`](https://github.com/messivite/goSupabase/blob/main/postman/README.md) for the same notes next to the file.
