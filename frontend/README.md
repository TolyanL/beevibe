# BeeVibe frontend

Vue 3 + TypeScript single-page app for the BeeVibe GameGen service. It talks to
the Go backend in `../backend` through the real HTTP API, including cookie auth,
SSE generation progress, playable iframes, patches, config updates, rollbacks,
repairs, playtime and the 1–5 rating scale.

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
  LLM key configured. Real generation needs `llm.api_key` in
  `backend/configs/config.yaml`.

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
bun run dev        # Vite dev server with HMR and backend proxy
bun run type-check # vue-tsc
bun run lint       # oxlint + ESLint (auto-fix)
bun run build      # type-check + production bundle into dist/
bun run preview    # preview the production build locally (same backend proxy)
bun run format     # oxfmt
```

## Production

```sh
bun run build
```

Serve `dist/` with an SPA fallback to `index.html`, and reverse-proxy `/api/*`
and `/g/*` to the Go backend on the same origin. Keeping the game iframe on the
same origin as the app is the default. If games are served from a separate
origin, set `VITE_GAME_BASE_URL` at build time (for example
`https://games.example.com`); the backend must then allow that origin for the
public `/g/*` routes.

## Routes

| Route            | Purpose                                                       |
| ---------------- | ------------------------------------------------------------- |
| `/`              | Generator: prompt, genre/style, map size, tempo, live preview |
| `/catalog`       | Public catalog with local search, genre filters and sorting  |
| `/login`         | Sign in                                                       |
| `/register`      | Sign up                                                       |
| `/profile`       | Balance, top-up, games and token operations                  |
| `/games/:id`     | Owner workspace: stage timeline, patches, versions, rating   |
| `/play/:id`      | Public playable page and shareable wrapper                    |
| `/g/:id`         | Backend-served self-contained HTML5 game (iframe)             |

`/games/*`, `/profile` and the generator's submit action require auth.
`/play/*` and `/catalog` are public.

## API integration

- `src/api/client.ts` — fetch wrapper, `ApiError`, error message mapping.
- `src/api/auth.ts` — register, login, logout, `me`, profile, top-up.
- `src/api/games.ts` — create, get, patch, config, rollback, fix, rate,
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

The public play page does not call the owner-only `GET /api/games/{id}`. It
loads the first 100 catalog cards to find metadata for the game, while the game
itself is always served from `/g/{id}`. Playtime is flushed to the backend only
for authenticated visitors.

## Regenerating API types

The backend contract is `../backend/api/http/openapi.yaml`:

```sh
bunx openapi-typescript ../backend/api/http/openapi.yaml -o src/api/schema.d.ts
```

The hand-written domain types in `src/api/types.ts` are the ones used by the UI;
`schema.d.ts` is the generated reference for the raw contract.

## Notes

- The theme is custom (cream light mode, near-black dark mode) and built on
  OpenVue's preset API plus Tailwind tokens. No CDN assets are loaded.
- The game iframe is sandboxed with scripts, pointer lock and modals enabled.
  Same-origin access, storage, form submission, popups and downloads are blocked
  so generated HTML cannot reach the app origin.
- The app uses native SVG icons and local `@fontsource` fonts only.
