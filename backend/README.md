# BeeVibe backend

Short reference for the service: what each endpoint accepts and returns.

- Machine-readable contract: `api/http/openapi.yaml`.
- Configuration: `configs/config.yaml` (template `configs/config.example.yaml`).
- Run and check commands: `Makefile` (`make run`, `make test`, `make lint`, `make mocks`).

Default port is `8080`, the API is mounted under `/api`, the public game pages
are mounted at the root. The service name is `beevibe`.

## Conventions

- **Auth** — a session cookie named `session` (HttpOnly, SameSite=Lax, 30 days).
  It is set by `register` and `login`, cleared by `logout`. Every route under
  `/api` except `health`, `catalog`, `auth/*` requires it and answers `401`
  without it.
- **Errors** — one shape for every failure:

  ```json
  { "error": { "code": "not_enough_tokens", "message": "not enough tokens" } }
  ```

- **Idempotency** — `POST /api/games` accepts an `Idempotency-Key` header. A repeat
  with the same key returns the same game and does not charge twice.

### Error codes

| Code                      | HTTP | Meaning                                                 |
| ------------------------- | ---- | ------------------------------------------------------- |
| `invalid_input`           | 400  | malformed body or a failed validation tag               |
| `validation_failed`       | 400  | the produced game did not pass the self-contained check |
| `unknown_config_key`      | 400  | the config block has a key the service does not know    |
| `unauthorized`            | 401  | no session, or the session is unknown/expired           |
| `forbidden`               | 403  | the game belongs to another user                        |
| `not_found`               | 404  | no such game                                            |
| `version_not_found`       | 404  | no such version of the game                             |
| `email_taken`             | 409  | the email is already registered                         |
| `not_enough_tokens`       | 409  | the balance does not cover the operation                |
| `generation_in_progress`  | 409  | the game is not in a state that allows the call         |
| `invalid_transition`      | 409  | the status machine refused the transition               |
| `idempotency_in_progress` | 409  | the same key is still being processed                   |
| `fix_limit_reached`       | 409  | the repair attempts are exhausted                       |

## Health

### `GET /api/health`

No auth.

```json
{ "status": "ok", "version": "0.2.0" }
```

## Auth

### `POST /api/auth/register`

```json
{ "email": "user@example.com", "password": "password123" }
```

`email` must be a valid address, `password` 8..72 characters.
**201** — user plus the session cookie:

```json
{ "user": { "id": "uuid", "email": "user@example.com", "balance": 1000000 } }
```

Errors: `400 invalid_input`, `409 email_taken`. The signup grant is configurable
(`tokens.grant_signup`).

### `POST /api/auth/login`

```json
{ "email": "user@example.com", "password": "password123" }
```

**200** — the same body as `register`, plus a fresh session cookie.
Errors: `401 unauthorized` for both a wrong password and an unknown email.

### `POST /api/auth/logout`

**204**, the session cookie is cleared. An unknown token is not an error.

## Profile

### `GET /api/me`

**200**

```json
{ "user": { "id": "uuid", "email": "user@example.com", "balance": 640000 } }
```

### `GET /api/profile`

**200** — balance, the last 50 games, the last 50 token operations and the
current prices (the client can show them without a second call):

```json
{
  "balance": 640000,
  "games": [
    {
      "id": "uuid",
      "prompt": "penguin arcade",
      "genre": "arcade",
      "status": "ready",
      "tokens_cost": 181081,
      "model_calls": 6,
      "fixes": 0,
      "created_at": "2026-10-23T22:33:00Z"
    }
  ],
  "token_ops": [
    {
      "game_id": "uuid",
      "delta": -320000,
      "reason": "generate",
      "created_at": "2026-10-23T22:33:00Z"
    }
  ],
  "prices": { "generate": 320000, "patch": 40000 }
}
```

`token_ops[].reason` — `signup_grant`, `generate`, `patch`, `topup`, `refund`.

### `POST /api/profile/topup`

No body. **200** — the new balance:

```json
{ "balance": 1140000 }
```

The amount comes from `tokens.grant_topup`.

## Games

### `POST /api/games`

Headers: `Idempotency-Key` (optional).

```json
{
  "prompt": "penguin arcade about catching fish",
  "genre": "arcade",
  "tempo": 5,
  "mapSize": "small",
  "style": "pixel"
}
```

| Field     | Rule                                                                     |
| --------- | ------------------------------------------------------------------------ |
| `prompt`  | required, 3..2000 characters                                             |
| `genre`   | optional: `arcade`, `snake`, `clicker`, `platformer`, `puzzle`, `custom` |
| `tempo`   | optional integer 1..10                                                   |
| `mapSize` | optional: `small`, `medium`, `large`                                     |
| `style`   | optional: `pixel`, `cartoon`, `neon`, `minimal`                          |

**201** — the game is queued and the generation price is already debited:

```json
{ "id": "uuid", "status": "queued" }
```

Errors: `400 invalid_input`, `409 not_enough_tokens`, `409 idempotency_in_progress`.

### `GET /api/games/{id}`

**200** — the current state, the stage timeline and the cost of every version:

```json
{
  "status": "ready",
  "stages": [{ "stage": "analyzing", "status": "done", "durationMs": 13495 }],
  "spec": {
    "genre": "arcade",
    "mechanic": "...",
    "controls": "...",
    "win_condition": "...",
    "lose_condition": "...",
    "palette": ["#101820"],
    "map_size": "small",
    "style": "pixel",
    "audience": "...",
    "assets": ["fish", "gull"]
  },
  "config": {
    "tempo": 5,
    "palette": ["#101820"],
    "map_size": "small",
    "difficulty": "normal",
    "controls": "keyboard+touch",
    "style": "pixel"
  },
  "versions": [
    {
      "id": "uuid",
      "number": 1,
      "source": "generate",
      "tokens_cost": 181081,
      "model_calls": 6,
      "created_at": "2026-10-23T22:33:00Z"
    }
  ],
  "cost": { "tokens": 181081, "calls": 6, "fixes": 0 },
  "rating": { "average": 4.6, "total": 18, "counts": [0, 1, 1, 4, 12] },
  "my_rating": 5
}
```

`stages[]` is the latest state per stage; `versions[].source` — `generate`,
`config`, `patch`, `repair`, `rollback`, `review`.

`rating.average` is the mean of all votes on a 1..5 scale, `rating.total` is the
number of votes, and `rating.counts` holds the number of 1, 2, 3, 4 and 5 star
votes in that order. `my_rating` is the owner's current vote, or `0` when the
owner has not rated the game yet.

Errors: `403 forbidden`, `404 not_found`.

### `GET /api/games/{id}/stream`

Server-Sent Events, `text/event-stream`. The stream replays the events it has and
then waits for new ones; it closes itself once the game is `ready` or `failed`
and nothing is left. A reconnect may send `Last-Event-ID` to continue after the
last seen sequence number. A `: ping` comment arrives every 15 seconds.

```
id: 1
event: stage
data: {"stage":"analyzing","status":"running","durationMs":0}

id: 2
event: artifact
data: {"kind":"spec","content":{...}}

id: 9
event: done
data: {"versionId":"uuid","cost":{"tokens":181081,"calls":6}}

id: 10
event: error
data: {"code":"validation_failed","message":"..."}
```

| Event      | Payload                                               |
| ---------- | ----------------------------------------------------- |
| `stage`    | `{stage, status: running\|done, durationMs, issues?}` |
| `artifact` | `{kind: spec\|config, content}`                       |
| `done`     | `{versionId, cost: {tokens, calls}}`                  |
| `error`    | `{code, message}`                                     |

Errors after the response has started arrive as an `error` event, not as an HTTP
status.

### `POST /api/games/{id}/patch`

```json
{ "text": "add a sprint on Shift" }
```

`text` — required, 2..4000 characters. The patch price is debited before the
model call and refunded if the result is rejected.

**200**

```json
{ "versionId": "uuid", "tokensSpent": 27128 }
```

Errors: `400 invalid_input`, `400 validation_failed`, `403 forbidden`,
`404 not_found`, `409 generation_in_progress` (the game is not `ready`),
`409 not_enough_tokens`.

### `PATCH /api/games/{id}/config`

Edits the `CONFIG` block of the latest version without the model, so it is free.
Every field is optional; omitted fields keep their value.

```json
{
  "tempo": 9,
  "palette": ["#101820", "#f2a541"],
  "mapSize": "medium",
  "difficulty": "hard"
}
```

`tempo` 1..10, `palette` up to 8 colours, `mapSize` `small|medium|large`,
`difficulty` up to 32 characters.

The request uses camelCase (`mapSize`), while the `config` object in
`GET /api/games/{id}` keeps the game-side names (`map_size`, `tempo`, `palette`,
`difficulty`, `controls`, `style`).

**200** — `{ "versionId": "uuid" }`.
Errors: `400 invalid_input`, `403 forbidden`, `404 not_found`,
`409 generation_in_progress`.

### `POST /api/games/{id}/rollback`

```json
{ "version": 1 }
```

`version` — the version number to restore, required, 1 or more. A new version is
created from the old HTML; nothing is charged.

**200** — `{ "versionId": "uuid" }`.
Errors: `400 invalid_input`, `403 forbidden`, `404 version_not_found`,
`409 generation_in_progress`.

### `POST /api/games/{id}/fix`

```json
{ "error": "jump does not work" }
```

`error` — optional, up to 2000 characters; the runtime message from the preview.
The service validates the game, asks the model to repair it and re-validates, up
to `worker.max_fix_attempts` times. Free.

**200** — `{ "versionId": "uuid" }`.
Errors: `400 invalid_input`, `403 forbidden`, `404 not_found`,
`409 generation_in_progress`, `409 fix_limit_reached`.

### `POST /api/games/{id}/rate`

```json
{ "value": 5 }
```

`value` — an integer from 1 to 5, where 5 is the best score. One vote per user,
a repeat overwrites the previous vote.

**200**

```json
{
  "rating": 5,
  "aggregate": {
    "average": 4.6,
    "total": 18,
    "counts": [0, 1, 1, 4, 12]
  }
}
```

`aggregate.average` is the mean score on the 1..5 scale, `aggregate.total` is the
number of votes, and `aggregate.counts` holds the counts for 1, 2, 3, 4 and 5
stars in that order.

Errors: `400 invalid_input`, `403 forbidden`, `404 not_found`.

### `POST /api/games/{id}/playtime`

```json
{ "seconds": 120 }
```

`seconds` — 0..300 per report; a larger value is rejected with `400 invalid_input`.

**200** — the total for the game: `{ "totalSeconds": 120 }`.
Errors: `400 invalid_input`, `404 not_found`.

## Catalog

### `GET /api/catalog?sort=rating&genre=arcade&limit=20&offset=0`

No auth. Only `ready` games are listed.

| Query    | Rule                                                                 |
| -------- | -------------------------------------------------------------------- |
| `sort`   | `rating` (average score), `hours` (playtime), `fresh` (default)      |
| `genre`  | the same enum as `POST /api/games`                                   |
| `limit`  | 1..100, defaults to 20                                               |
| `offset` | 0 or more                                                            |

**200**

```json
{
  "items": [
    {
      "id": "uuid",
      "prompt": "penguin arcade",
      "genre": "arcade",
      "map_size": "small",
      "style": "pixel",
      "created_at": "2026-10-23T22:33:00Z",
      "rating": { "average": 4.6, "total": 18, "counts": [0, 1, 1, 4, 12] },
      "playtime": { "sessions": 1, "seconds": 120, "hours": 0.03 },
      "tokens_cost": 181081,
      "model_calls": 6
    }
  ]
}
```

Errors: `400 invalid_input` for an unknown `sort`/`genre`.

## Public pages

Mounted at the root, not under `/api`.

### `GET /g/{id}`

**200** `text/html` — the latest version of the game as a self-contained document
(no external resources, works offline from the filesystem).
Errors: `400` for a malformed id, `404 not_found`.

### `GET /g/{id}/download`

The same document with `Content-Disposition: attachment; filename="game.html"`.

## Notes

- Prices and grants come from `configs/config.yaml` (`tokens.*`), the effort grid
  per stage from `llm_reasoning.*`; the API does not expose them except through
  `GET /api/profile`.
- A failed generation refunds the full price (`token_ops.reason = refund`).
- Lifecycle events (`game.created`, `game.ready`, `game.failed`, `game.patched`)
  are published to the Kafka topic `game.events` through the outbox; they are not
  part of the HTTP contract.
