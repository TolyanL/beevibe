# BeeVibe

![BeeVibe](assets/brand/beevibe-cover.jpg)

**A playable game from one sentence.** Describe an idea in a sentence — a pipeline of
LLM agents turns it into a self-contained HTML5 game you can play in the browser,
patch with words, rate and publish to a catalog.

Built as the Beeline case solution for the **ITSprint 2026** hackathon.

```
"a neon snake in a cyberpunk city, the tempo grows with the tail"  ->  /g/neon-snake  ->  play
```

## What it does

- **Generate** — prompt plus optional title, description, genre, style, tempo and map
  size produce a spec, a plan, code and assets, assembled into one HTML file.
- **Watch it build** — live progress over SSE: stage timeline, artifacts, terminal log,
  refund-aware errors.
- **Iterate** — patch with plain language, tune tempo/map/difficulty/palette, repair a
  broken build, roll back to any version.
- **Name it** — title and description can be set at creation or edited at any time;
  a paid action lets the model invent the name, description and cover from the build.
- **Publish** — a public page with a sandboxed iframe, a shareable link, a download
  button and a 1–5 rating.
- **Catalog** — search, genre filters, sorting by freshness, rating or playtime, plus
  time-played tracking.
- **Tokens** — signup grant, per-action prices, demo top-up with a daily cap, full
  operation history. A failed build is refunded.

## How the pipeline works

Seven agent stages run in a worker pool; every stage writes an artifact to Postgres, so
a rerun can resume from where it stopped instead of paying for the whole pass again.

Two speed modes are available at creation time (`speed` in the request):

| Mode | Calls | Pipeline | Price |
| --- | --- | --- | --- |
| `normal` | 7 | analyze → plan → code → assets → review → assemble → check | full |
| `fast` | 2 | spec+plan → the whole game as one HTML document → check | −25% |

The fast mode keeps the same status vocabulary (`analyzing`, `coding`, `checking`) as a
subsequence of the full pipeline, so resume, retry and the progress UI work in both modes.
It has no review stage, one repair attempt and its own LLM client (`fast.llm`), and it
streams the big generation call so a provider gateway does not cut it off.

```
prompt
  |
  v
[1] analyze  -> genre, mechanics, constraints
[2] plan     -> rules, win/lose conditions, structure
[3] code     -> the game itself
[4] assets   -> palette, visual language
[5] review   -> runs the HTML, compares it with the plan
      |            \
      |             `-> needs_decision: the author either accepts the code
      |                 or asks for the code stage only
[6] assemble -> one self-contained HTML document
[7] check    -> final validation (self-contained, JS syntax)
  |
  v
version ready -> playable at /g/<slug>
```

In the fast mode the same path is compressed to two calls:

```
prompt
  |
  v
[1] spec+plan  -> spec JSON and the plan in one answer
[2] one-shot   -> the whole self-contained game in one streamed HTML answer
  |
  v
check -> version ready
```

Failure at any stage marks the game `failed`; the author can repair it with a hint or
retry from the failed stage. Editing operations (patch, config, repair, rollback)
produce new versions on top of the same game, so the owner always keeps the history;
naming (`PATCH /meta`) only touches the card.

## Architecture

```
                    browser
        +--------------+---------------+
        |                              |
   Vue 3 SPA                     sandboxed iframe
   (Vite, Pinia, Tailwind)       (generated /g/<slug>)
        |                              ^
        | REST + SSE, session cookie   |
        v                              |
+--------------------+                 |
|  Go service        |  serves the SPA |
|  chi router        |  from http.static_dir
|  usecase / adapter |
+---------+----------+
          |                    |
          v                    v
     PostgreSQL            Kafka topic game.events
   (games, versions,       (outbox -> consumer ->
    artifacts, tokens,      game_event_log projection)
    ratings, sessions)
          |
          v
    LLM provider (OpenRouter or any OpenAI-compatible server)
```

- **Backend** — layered Go service (`domain -> dto -> usecase -> adapter/controller`),
  transactions wrapped explicitly, outbox for Kafka, ids and slugs carried in the API.
- **Frontend** — Vue 3 SPA with a custom cream/near-black theme, SSE workspace,
  sandboxed playable iframe, ratings and catalog.
- **One image** serves both: the Go binary serves the built SPA (`http.static_dir`) and
  the API on the same origin, so cookies, SSE and iframes need no CORS setup.

## Stack

| Layer      | Technology                                                                 |
| ---------- | -------------------------------------------------------------------------- |
| Backend    | Go 1.27, chi v5, pgx v5, segmentio/kafka-go, zerolog, go-playground/validator, goose migrations |
| Database   | PostgreSQL 18 (games, versions, artifacts, events, outbox, tokens, ratings) |
| Messaging  | Kafka (KRaft) with an outbox table and an event-log projection             |
| LLM        | OpenRouter or any OpenAI-compatible endpoint (llama.cpp, vLLM); reasoning effort per stage |
| Frontend   | Vue 3 + TypeScript, Vite 8, Pinia, Vue Router 5, OpenVue (PrimeVue 4 fork), Tailwind CSS 4, Bun |
| Deployment | Multi-stage Docker image (Bun build + Go build + Alpine runtime with goose), Docker Compose stack |
| CI         | GitHub Actions: gofmt/vet/golangci-lint/tests for the backend, oxfmt/oxlint/ESLint for the frontend |

## Repository layout

```
backend/               Go service
  cmd/app              entrypoint
  internal/domain      entities, invariants, status machine
  internal/usecase     operations, ports, worker pool, pipeline
  internal/adapter     postgres, kafka, auth (bcrypt), validator
  internal/controller  http v1 + public pages + kafka consumer
  internal/dto         request/response shapes
  api/http/openapi.yaml  the contract the frontend and READMEs follow
  migrations/          goose migrations
  configs/             config.example.yaml (committed), config.yaml (local)
  pkg/                 httpserver, static (SPA), llm client, render, logger,
                       transaction, tracer, postgres
frontend/              Vue 3 SPA
  src/pages            generator, catalog, workspace, play, auth, profile
  src/components       layout, catalog, game frame, generator preview
  src/stores           auth, catalog, game (SSE lifecycle), theme
  src/api              typed client, SSE wrapper, generated schema
deploy/                container config template and sync script
docker-compose.yml     app + Postgres + Kafka + migrations stack
docker-compose.dev.yaml  local infrastructure for development
Dockerfile             single image with the SPA and the API
```

## Getting started

### Option 1 — the container stack

Requirements: Docker with Compose. No script and no local Go/Bun toolchain — the stack
runs the published image.

```sh
# 1. the config the container reads (gitignored), derived from your local one
cp deploy/config.example.yaml deploy/config.yaml   # then fill llm.api_key
#    or keep one source of truth with the local dev config:
./deploy/sync-config.sh

# 2. pull the published image and run app + Postgres + Kafka + migrations
APP_PORT=8080 docker compose up -d --wait
```

`docker compose up -d` uses `ghcr.io/tolyanl/beevibe:latest`; add `--build` to build from
the working tree instead, or pin a version with
`IMAGE=ghcr.io/tolyanl/beevibe:v1.4.0 docker compose up -d`. `--wait` blocks until the
`/api/health` healthcheck passes.

The app lands on `http://localhost:8080`: the SPA at `/`, the API under `/api`,
generated games at `/g/<slug>`. Migrations run as a one-shot service before the app
starts, `goose` is inside the image. Kafka creates `game.events` on the first write
(`KAFKA_NUM_PARTITIONS: 8`), so no topic-init container is needed.

`deploy/config.yaml` is not committed: it holds the LLM key, so it never goes to git —
start from `config.example.yaml` or `./deploy/sync-config.sh` instead. Only three values
differ from the template — log format, the Postgres and Kafka hosts inside the compose
network. The file has to exist **before** `docker compose up`: Docker bind-mounts a
directory for a missing path, so the app checks for the file and exits with
`deploy/config.yaml is missing: run ./deploy/sync-config.sh` instead of a cryptic
`is a directory`. `docker compose exec app …` works for debugging; the health endpoint is
`/api/health`.

### Option 2 — local development

Requirements: Go 1.27, Bun, Docker (for Postgres and Kafka only).

```sh
# infrastructure: Postgres + Kafka + Kafka UI
docker compose -f docker-compose.dev.yaml up -d

cd backend
make config          # creates configs/config.yaml from the example
make migrate-up      # goose migrations
make run             # the service on :8080

cd ../frontend
bun install
bun run dev          # SPA on :5173, proxying /api and /g to :8080
```

Fill `llm.api_key` in `backend/configs/config.yaml` to generate games. Any
OpenAI-compatible endpoint works:

```yaml
llm:
  api_key: "sk-or-..."                     # or any string for a local server
  model: deepseek/deepseek-v4.1-flash-1017 # e.g. a local qwen build
  base_url: https://openrouter.ai/api/v1   # e.g. http://127.0.0.1:12286/v1
  reasoning_dialect: openrouter            # or "openai" for llama.cpp / vLLM
```

### Option 3 — the published image

#### The config is mandatory — the container cannot start without it

The image carries only `configs/config.example.yaml`. The binary reads `APP_CONFIG_PATH`,
which is baked into the image as `/app/configs/config.yaml`, so **the config has to live
at `/app/configs/config.yaml` inside the container**, mounted read-only from a file on the
host:

| On the host                 | In the container               | Why                                        |
| --------------------------- | ------------------------------ | ------------------------------------------ |
| `deploy/config.yaml`        | `/app/configs/config.yaml:ro`  | the file you edit; gitignored, holds the key |

```sh
cp deploy/config.example.yaml deploy/config.yaml   # once, then fill llm.api_key
#    or derive it from your local dev config, hosts rewritten for the network:
./deploy/sync-config.sh
```

Without the mount the container dies immediately:

```json
{"level":"fatal","error":"os.ReadFile: open /app/configs/config.yaml: no such file or directory","message":"config.New"}
```

The file is read once at startup — after editing it, recreate the container
(`docker restart beevibe_app` is enough).

#### Prerequisites: network, database, migrations

Everything is in `docker-compose.yml` — `docker compose up -d` creates the network, starts
Postgres and Kafka, applies the migrations and then the app (Option 1). The
`beevibe_default` network only exists while that compose project has containers, so the
`docker run` below is for recreating just the app container on a running stack.

#### Run only the app container

```sh
# stack already up (docker compose up -d db kafka migrate)
docker run -d --name beevibe_app --network beevibe_default -p 8080:8080 \
  -v "$PWD/deploy/config.yaml:/app/configs/config.yaml:ro" \
  ghcr.io/tolyanl/beevibe:v1.4.0

# local model on the host: reach it through host.docker.internal
docker run -d --name beevibe_app --network beevibe_default -p 8080:8080 \
  -v "$PWD/deploy/config.yaml:/app/configs/config.yaml:ro" \
  --add-host host.docker.internal:host-gateway \
  ghcr.io/tolyanl/beevibe:v1.4.0
```

`$PWD` must be the repository root, where `deploy/config.yaml` lives; use an absolute
path (`-v /srv/beevibe/deploy/config.yaml:…`) if you start the container from elsewhere.

Check that it is really up — the SPA at `/` answers 200 even with a dead database, so
trust the API, not the page:

```sh
curl -s localhost:8080/api/health      # {"status":"ok","version":"1.4.0"}
curl -s localhost:8080/api/catalog     # {"items":[],"total":0}
```

#### What each mistake looks like

| Symptom | Cause |
| ------- | ----- |
| `failed to set up container networking: network beevibe_default not found` | the compose stack is not running on this machine — `docker compose up -d` creates the network |
| container exits, `pool.Ping: failed to connect to … lookup db` | no `--network beevibe_default` (or `db`/`kafka` still point at the compose hosts) — the DSN host and the broker are only resolvable inside the network |
| container runs, `/` answers 200, every API call fails with `lookup db` | same cause on the older tags that did not ping the database at startup; the page proves nothing |
| `os.ReadFile: open /app/configs/config.yaml: no such file or directory` | the `-v …:/app/configs/config.yaml:ro` mount is missing |
| `port is already allocated` | another container publishes 8080 — stop it or pass `-p 8081:8080` |

#### Local model (llama.cpp/vLLM on the host)

In `deploy/config.yaml`: `base_url: http://host.docker.internal:12286/v1`,
`reasoning_dialect: openai`, `timeout: 3600s`; start with `--add-host` (command above).

#### Model through an API (OpenRouter and any OpenAI-compatible)

In `deploy/config.yaml`:

```yaml
llm:
  api_key: "sk-or-..."
  model: deepseek/deepseek-v4.1-flash-1017
  base_url: https://openrouter.ai/api/v1
  reasoning_dialect: openrouter
  timeout: 600s
  max_tokens: 0
  max_retries: 1
```

No `--add-host` needed: outgoing internet access and a valid key are.
`timeout: 600s` is enough for an API, unlike `3600s` for a local model: the
per-attempt timeout has to outlive the model's thinking, which runs for minutes
before the first content token. `max_tokens: 0` leaves the output uncapped: a
reasoning model spends ~20k tokens thinking before writing, so a tighter limit
returns an empty answer with `finish_reason: length`.

## Configuration

`backend/configs/config.yaml` (`APP_CONFIG_PATH` overrides the path). Missing fields fall
back to defaults in `internal/config`; a mandatory parameter is validated on startup.

| Section          | What lives there                                                                 |
| ---------------- | -------------------------------------------------------------------------------- |
| `app`            | name and version reported by `/api/health`                                       |
| `http`           | port, timeouts (`write_timeout: 0s` keeps SSE open), `static_dir` for the SPA     |
| `postgres`       | DSN and pool settings                                                            |
| `llm`            | provider, model, timeout/max_tokens/retries, reasoning dialect, debug dump directory |
| `llm_reasoning`  | effort per stage: `default`, `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max` |
| `tokens`         | signup grant, prices (generate/patch/meta), demo top-up limits                   |
| `worker`         | concurrency, per-game timeout, call budget, repair attempts                      |
| `kafka`          | brokers, topic (`game.events`), consumer group                                   |
| `auth`           | bcrypt cost                                                                      |
| `validator`      | optional Node path for the generated-JS syntax check                             |

Tokens: a new account gets `tokens.grant_signup` (1,000,000 by default), a full build
costs `price_generate` (320,000), a patch 40,000, model-written meta 10,000. Demo top-up
is limited per request and per day.

## API

The machine-readable contract is [`backend/api/http/openapi.yaml`](backend/api/http/openapi.yaml);
the endpoint-by-endpoint reference with examples is [`backend/README.md`](backend/README.md).
The frontend is documented in [`frontend/README.md`](frontend/README.md).

Highlights:

| Method | Path                            | What it does                                    |
| ------ | ------------------------------- | ----------------------------------------------- |
| POST   | `/api/games`                    | create a game (debits tokens, queues the build)  |
| GET    | `/api/games/{id}/stream`        | SSE: stages, artifacts, done, error              |
| POST   | `/api/games/{id}/patch`         | change the game with words                       |
| PATCH  | `/api/games/{id}/config`        | tempo/map/difficulty/palette without the model   |
| PATCH  | `/api/games/{id}/meta`          | title, description and cover (free, any status)  |
| POST   | `/api/games/{id}/meta/generate` | model-written name, description and cover        |
| PATCH  | `/api/games/{id}/visibility`    | publish to the catalog or hide                   |
| GET    | `/api/catalog`                  | public catalog with search, filters and sorting  |
| GET    | `/g/{id}`                       | the game as self-contained HTML                  |

Conventions: a `session` cookie (HttpOnly) for auth, one error shape
(`{"error": {"code", "message"}}`), `Idempotency-Key` on mutating calls that publish
events, and domain errors mapped to HTTP codes at the boundary.

## Releasing the image

`.github/workflows/docker.yaml` builds the `Dockerfile` and pushes
`ghcr.io/tolyanl/beevibe` on every `v*` tag (`workflow_dispatch` with a tag input does the
same by hand):

```sh
# 1. bump app.version everywhere (config.example.yaml, config.go Defaults, openapi,
#    frontend/package.json) — the workflow refuses a tag that disagrees with it
# 2. tag and push
git tag v1.4.0 && git push origin v1.4.0
```

Published tags per release: `v1.4.0`, `1.4.0`, `1.4` and `latest`. The job logs into ghcr
with `GITHUB_TOKEN` (`packages: write`), so no PAT or repository secret is needed — the
package only has to be linked to the repository once (Settings → Packages). Nothing is
pushed until the build succeeds, and the `app.version` ↔ tag check keeps the tag,
`app.version` and the `IMAGE=` override of `docker compose` in sync.

## Quality gates

```sh
cd backend
make lint            # golangci-lint
make test            # go test -v -cover ./...   (no database, broker or LLM needed)
go build ./...       # what CI runs as well

cd frontend
bun run type-check   # vue-tsc
bun run check        # oxlint + ESLint, zero warnings
bun run format:check # oxfmt
bun run test         # vitest: api layer and page behaviour
bun run build        # type-check + production bundle
```

Tests never require infrastructure: Postgres, Kafka and the LLM are replaced by mocks or
`httptest` servers, and adapters take their executor through a context seam.

## Notes and known limits

- Generated games are sandboxed iframes (`allow-scripts`, no same-origin), so a build
  cannot touch the app session or storage.
- The catalog only lists `ready` + `public` games; publishing is owner-only.
- The idempotency key expires after an hour; a failed operation blocks a retry with the
  same key until then (operator can delete `idempotency:<key>`).
- Real generation needs a working LLM: without a key the service starts and serves the
  catalog, but `/api/games` requests fail at the first model call.
- Playtime is flushed only for authenticated visitors; ratings are stored per account.
