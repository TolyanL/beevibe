# BeeVibe frontend

Vue 3 + TypeScript single-page app for the BeeVibe service. It talks to the Go
backend in `../backend` through the real HTTP API: cookie auth, SSE generation
progress, playable iframes, patches, config updates, rollbacks, repairs, game
naming, playtime and the 1–5 rating scale.

## Stack

- Vue 3 with `<script setup>` and TypeScript
- Vite 8
- Pinia
- Vue Router 5
- OpenVue (`openvue`, the PrimeVue 4 fork) and `@openvue/themes`
- Tailwind CSS 4 with `tailwindcss-primeui`
- `openapi-typescript` generated types in `src/api/schema.d.ts`

## Requirements

- Bun 1.x (or a compatible Node.js 22.18+/24.12+ runtime)
- The backend running on `http://127.0.0.1:8080` with PostgreSQL, Kafka and the
  LLM configured. Real generation needs `llm.api_key` (and a reachable
  `llm.base_url` — OpenRouter or any OpenAI-compatible server such as llama.cpp)
  in `backend/configs/config.yaml`.

## Development

```sh
bun install
bun run dev
```

The app opens on `http://127.0.0.1:5173`.

Vite proxies same-origin requests to the backend:

| Path    | Target                                      |
| ------- | ------------------------------------------- |
| `/api/*` | `VITE_BACKEND_URL` (default `http://127.0.0.1:8080`) |
| `/g/*`   | `VITE_BACKEND_URL`                          |

The proxy keeps the session cookie and SSE on the same origin, so no CORS setup
is needed on the backend. `/games/*` stays inside the SPA for the owner
workspace; only `/g/*` is proxied.

To point the dev proxy at another backend:

```sh
VITE_BACKEND_URL=http://127.0.0.1:9090 bun run dev
```

## Scripts

```sh
bun run dev          # Vite dev server with HMR and backend proxy
bun run type-check   # vue-tsc
bun run lint         # oxlint + ESLint with auto-fix
bun run check        # oxlint + ESLint, zero warnings (what CI runs)
bun run format       # oxfmt
bun run format:check # oxfmt in check mode
bun run build        # type-check + production bundle into dist/
bun run build-only   # the bundle without the type-check
bun run preview      # preview the production build locally (same backend proxy)
```

## Production

The Go service serves the bundle itself: the image from the repository root copies
`dist/` into `/app/static`, and `http.static_dir` in `backend/configs/config.yaml`
points there (`static_dir: static`). Unknown paths fall back to `index.html`, so deep
links work; `/api/*` and `/g/*` stay on the same origin, which keeps the session
cookie, SSE and the game iframe happy without CORS.

```sh
# from the repository root: builds the SPA and the API into one image
docker compose up -d --build
```

Serving `dist/` from a separate web server is still possible — reverse-proxy
`/api/*` and `/g/*` to the Go backend on the same origin. If games are served from a
different origin, set `VITE_GAME_BASE_URL` at build time (for example
`https://games.example.com`); the backend must then allow that origin for the public
`/g/*` routes.

## Routes

| Route            | Purpose                                                       |
| ---------------- | ------------------------------------------------------------- |
| `/`              | Generator: name, description, prompt, genre/style, map, tempo, live catalog-card preview |
| `/catalog`       | Public catalog with local search, genre filters and sorting  |
| `/login`         | Sign in                                                       |
| `/register`      | Sign up                                                       |
| `/profile`       | Balance, top-up, games and token operations                  |
| `/games/:slug`   | Owner workspace: stage timeline, patches, versions, rating   |
| `/play/:slug`    | Public playable page and shareable wrapper                    |
| `/g/:slug`       | Backend-served self-contained HTML5 game (iframe)             |

`/games/*`, `/profile` and the generator's submit action require auth.
`/play/*` and `/catalog` are public.

## API integration

- `src/api/client.ts` — fetch wrapper, `ApiError`, error message mapping.
- `src/api/auth.ts` — register, login, logout, `me`, profile, top-up.
- `src/api/games.ts` — create, get, patch, config, rollback, fix, meta, rate,
  playtime.
- `src/api/catalog.ts` — public catalog.
- `src/api/sse.ts` — EventSource wrapper for `stage`, `artifact`, `done` and
  `error` events with native reconnect.
- `src/stores/game.ts` — the central workspace store; owns the SSE lifecycle,
  refresh key used to reload the iframe after a new version, patch/config/
  rollback/repair actions and the rating aggregate.

The 1–5 rating is real backend state, not a frontend fake. `GET /api/games/{id}`
returns `rating` (`average`, `total`, `counts`) and `my_rating`; `POST
/api/games/{id}/rate` accepts `{ "value": 1..5 }`. The UI colors the scale from
bright red at 1 through yellow at 3 to bright green at 5.

The public play page does not call the owner-only `GET /api/games/{id}`. It reads
the card from `GET /api/games/{slug}/public` (rating, playtime, `my_rating`,
`is_owner`), while the game itself is always served from `/g/{slug}`. Playtime is
flushed to the backend only for authenticated visitors.

## Regenerating API types

The backend contract is `../backend/api/http/openapi.yaml`:

```sh
bunx openapi-typescript ../backend/api/http/openapi.yaml -o src/api/schema.d.ts
```

The hand-written domain types in `src/api/types.ts` are the ones used by the UI;
`schema.d.ts` is the generated reference for the raw contract.

## Meta and naming

- The generator asks for an optional name and description before the prompt; both
  are stored with the game on creation (`POST /api/games`) and shown on the
  catalog card from the start.
- The owner workspace edits them at any time (`PATCH /api/games/{id}/meta`, free,
  works while the game is still building) and can publish the card.
- Model-written name, description and cover are a paid action available only for a
  built game (`POST /api/games/{id}/meta/generate`); it lands in the SSE stream as
  an `artifact` event with `kind: "meta"`.
- Renaming does not change the slug, so public links stay stable.

## Notes

- The theme is custom (cream light mode, near-black dark mode) and built on
  OpenVue's preset API plus Tailwind tokens. No CDN assets are loaded.
- In `src/theme/preset.ts` the dark `surface` scale runs from the lightest
  (`surface.0`) to the darkest (`surface.950`), matching what OpenVue expects in
  both schemes. Reversing it makes dialogs and popups render light while their
  text stays light, and toast details pick `surface.0` as a near-black colour.
- The game iframe is sandboxed with scripts, pointer lock and modals enabled.
  Same-origin access, storage, form submission, popups and downloads are blocked
  so generated HTML cannot reach the app origin.
- The app uses native SVG icons and local `@fontsource` fonts only.
