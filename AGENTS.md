# AGENTS.md — Go service: architecture and style

Rules for all code in this repository: a new service, a new feature, refactoring, review.

Priority: the actual linter (the command from `Makefile`/`Justfile`, §0) > this file > taste. If a rule disagrees with the linter, the linter is right, and the discrepancy is fixed by editing this file, not by adding `//nolint`.

Deviating from this file is either a bug or a reason to update this file. Silent deviation is not allowed; when you deviate, say so explicitly.

**This file pins neither the stack nor the Go version.** It describes architecture and style that apply to any stack. Mentions of libraries (validator, envconfig, zerolog, chi, pgx, mockery, etc.) describe a typical stack, not a mandatory one: the actual stack and Go version come from the repository (`go.mod`/`go.work`, actual imports, the commands file). The rules of §16 are applied according to the Go version in `go.mod`.

## 0. Two modes: a new repository and an existing service

### 0.1 A new (clean) repository — bootstrap

A clean repository has no code, no `Makefile`, and no `.golangci.yml`. Before generating code, **demand from the user**:

1. The linter config — `.golangci.yml` (the user's own, or explicit permission to take the standard linter set for such a service).
2. The run-and-check commands file — `Makefile` **or** `Justfile` — with a set of targets: build (`run`), lint (`lint`), unit tests (`test`), integration tests, migrations, generation. Target names are whatever the user is used to.
3. Service parameters: module path (`go.mod`), binary name, ports, stack (HTTP/gRPC/Kafka/Postgres/Redis...).

Rules:

- No code is written until the linter config and the commands file are in hand. Commands are not invented: only the ones the user gave.
- If the user says "create it yourself" — propose a standard `.golangci.yml` and `Makefile` with the canonical targets from §0.1 and get explicit confirmation before generating.
- The skeleton is created once, per §1: `cmd/`, `config/`, `internal/{app,domain,dto,usecase,adapter,controller}`, `pkg/`, `migration/`, `test/integration/`, `api/` (if the stack has a transport).
- The commands obtained are recorded in the "Project additions" section (the "Commands" line) and only they are used from then on.
- Before writing Go code, load the `go-style` skill — it holds the canonical examples and the language style; the rules live in this file.

### 0.2 An existing service

- Read `Makefile`/`Justfile`, `.golangci.yml`, `internal/app/app.go`. They describe the actual build, the actual linter set, and the actual dependency graph — they outrank any assumption. If there is no commands file, ask which commands to use for lint/tests/build, and do not invent them.
- A new operation → follow the §13 checklist from top to bottom: domain → dto → usecase → adapter → controller → test.
- Before handing the work over, the following are mandatory: the lint command (§0), `go build ./...`, tests — using the commands from the commands file.
- Touch only the layer the task concerns. Rewriting neighbouring layers "while you're at it" is forbidden — that is a separate task.
- Before writing Go code, load the `go-style` skill — it holds the canonical examples and the language style; the rules live in this file.

## 1. Layers and dependency direction

| Layer        | Path                     | May import                                                              |
| ------------ | ------------------------ | ----------------------------------------------------------------------- |
| `cmd`        | `cmd/<binary>`           | `internal/config`, `internal/app`, `pkg/*`                              |
| `config`     | `internal/config`        | the layers' `Config` types, `pkg/*`                                     |
| `app`        | `internal/app`           | everything from `internal/*`, `pkg/*`                                   |
| `controller` | `internal/controller/**` | `usecase`, `dto`, `domain`, `gen/*`, `pkg/*`                            |
| `usecase`    | `internal/usecase`       | `domain`, `dto`, `pkg/*`                                                |
| `adapter`    | `internal/adapter/**`    | `domain`, `dto`, `pkg/*`                                                |
| `domain`     | `internal/domain`        | stdlib, the project's domain libraries; DB drivers and transport — no   |
| `dto`        | `internal/dto`           | `domain`                                                                |
| `pkg`        | `pkg/**`                 | stdlib, `pkg/*`                                                         |
| `gen`        | `gen/**`                 | generated — never edited by hand                                        |

Always forbidden:

- importing `internal/*` from `pkg/*`;
- reading the service configuration from the environment: it lives in `configs/config.yaml`;
- `usecase` → `adapter`, `adapter` → `usecase`, `adapter` → `controller`;
- `domain` → `dto`, `pkg/*`, DB drivers, transport;
- any cyclic dependency;
- editing `gen/**` and `bin/**` by hand — regenerate instead (the commands from `Makefile`/`Justfile`; if there are no generation commands, ask the user).

## 2. Where to put a new package

The criterion is exactly one: **can this package be copied into another of my projects without a single change?**

- Yes → `pkg/<name>`.
- No (it knows about our domain, our service, our tables, our endpoints) → `internal/<...>`.

New code goes into `internal/` by default. Moving it to `pkg/` is a separate, deliberate decision. Clients of our services that know about our contract/domain belong in `internal/`, not `pkg/`.

`pkg/` holds only generic infrastructure: `httpserver`, `logger`, `metrics`, `otel`, `postgres`, `redis`, `transaction`, `render`, `router`.

## 3. Layer contract

### 3.1 domain (`internal/domain`)

- Entities with their own invariants and behaviour: `Profile`, `Property`, `Event`, `Status`.
- Typed primitives instead of "bare" strings and numbers: `type Name string`, `type Age int`.
- Creation only through a constructor that validates and returns an error: `NewProfile(...) (Profile, error)`.
- Mutation only through domain methods, validated as well (`Validate()` inside). Assigning fields from outside is forbidden (§6).
- Business methods, not "a struct with data": `IsDeleted()`, `Validate()`, `ToEvent(topic)`, `Status.String()`.
- Errors are singletons in a `var` block only: `ErrNotFound = errors.New("not found")`. Dynamic `errors.New`/`fmt.Errorf` for domain errors is forbidden (`err113`).
- `context` is not passed into domain. Domain knows nothing about SQL, HTTP, gRPC, Kafka.
- The global validator `var validate = validator.New(...)` is a sanctioned exception to the ban on global state and must carry `//nolint:gochecknoglobals`. The validation library is per project (typically `go-playground/validator`).

### 3.2 dto (`internal/dto`)

- `<Entity><Action>Input` / `<Entity><Action>Output`, one file per operation, snake_case.
- Fields are primitives or domain types; `json` tags are mandatory.
- `Output` may contain a domain entity (`Profile domain.Profile`) — that is allowed; there is no need to duplicate its fields as primitive fields.
- `Validate()` — if the input has shape constraints: pagination, `oneof` for sorting, UUID format. Do not duplicate entity invariants here (§6).
- dto contains no logic beyond `Validate()`. No access to the DB, the logger, or transport.

### 3.3 usecase (`internal/usecase`)

- Ports are declared here, in `usecase.go`, as interfaces: `Postgres`, `Redis`, `Kafka`. Interfaces are not created in other packages.
- `UseCase` + `New(postgres Postgres, redis Redis, kafka Kafka) *UseCase`. Fields are private and named after the port.
- One file per operation: `create_profile.go`, `outbox_read_and_produce.go`.
- Order inside an operation:

```go
func (u *UseCase) CreateProfile(ctx context.Context, input dto.CreateProfileInput) (dto.CreateProfileOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase CreateProfile")
	defer span.End()

	var output dto.CreateProfileOutput

	// 1. validate the input shape
	// 2. idempotency, if the operation mutates and publishes an event (§7)
	// 3. domain: only domain constructors and methods
	// 4. writes: only transaction.Wrap (§4)
	// 5. side effects: only after a successful commit (§8)

	return output, nil
}
```

- usecase does not write SQL, does not import `adapter`/`controller`/`gen`, and knows nothing about HTTP statuses or proto types.
- It always returns the error. `log.Error` + swallowing the error is forbidden (§5).
- It does not assign domain entity fields — only domain methods (§6).

### 3.4 adapter (`internal/adapter/<tech>`)

- Implements a port from usecase. One file per operation plus the package constructor file (`postgres.go`, `redis.go`).
- Constructor: `New(...) *Postgres`, dependencies as struct fields.
- SQL is a backtick constant, arguments are a separate slice:

```go
const sql = `SELECT created_at, name FROM profile WHERE id = $1`

txOrPool := transaction.TryExtractTX(ctx)

err := txOrPool.QueryRow(ctx, sql, profileID).Scan(dest...)
```

- `transaction.TryExtractTX(ctx)` is the only way to reach the DB. Calling it outside `internal/adapter/**` is forbidden (§4).
- For `Scan`, use a DTO struct in the same file with `Dest() []any` and `ToDomain()`, with driver types for nullable fields (typically `pgtype.*` for pgx v5).
- A driver's "no rows" (typically `pgx.ErrNoRows`) → a domain error (`domain.ErrNotFound`), not `nil` and not a driver error escaping outwards.
- After `Exec` (UPDATE/DELETE), check the driver's `RowsAffected` result (typically `CommandTag.RowsAffected()`). `errors.Is(err, pgx.ErrNoRows)` after `Exec` is a dead check and is not allowed (§14).
- SQL is not assembled with `fmt.Sprintf` from client-supplied values. A dynamic `ORDER BY`/column name may only come from an allow-list verified in `dto.Validate()`.
- The adapter does not decide the fate of the operation: it returns the error upwards; it has nothing to log and "carry on" with.

### 3.5 controller (`internal/controller/{http,grpc,kafka_consumer,worker}/v1`)

- A `Handlers` struct + `New(uc *usecase.UseCase) *Handlers` — the concrete usecase type, not an interface (§15).
- The controller does exactly three things: request → `dto.XxxInput`, call the usecase, `dto.XxxOutput` → response. There are no business decisions in the controller.
- An error is handled via `render.Error(ctx, err, "request failed")`, then the domain error is mapped to a code:

```go
output, err := h.usecase.CreateProfile(ctx, input)
if err != nil {
	err = render.Error(ctx, err, "request failed")

	return http_server.CreateProfile400JSONResponse{Error: err.Error()}, nil
}
```

- **The second return value is always `nil` when a response has been formed, 4xx included.** Returning `err` together with a response produces a 500: the strict handler calls `ResponseErrorHandlerFunc` (`http.StatusInternalServerError`) and the response body is lost.
- Domain error mapping via `errors.Is`: `ErrNotFound` → 404, `ErrIdempotencyInProgress`/conflict → 409, everything else → 400. Error strings are never compared.
- HTTP routing and middleware live in `internal/controller/http/router.go`: tracing, logger, metrics, then the API version (router and middleware per project, typically chi/v5).
- gRPC — `google.golang.org/grpc/status` + `codes.*`, server and interceptors in `grpc.go`.
- A worker (`worker/`) is a background loop with `stop`/`done` channels, released via `Close()`; here `log.Error` + continuing the loop is allowed (§5).

### 3.6 app, config, cmd

- `internal/app/app.go` is the single composition root: all implementations are created here, `usecase.New(...)` is wired here, and the shutdown order is defined here. No logic beyond assembly.
- Creation order: pool → clients → metrics → usecase → controllers → workers. Shutdown is strictly the reverse: controllers first, then adapters.
- `internal/config/config.go` is a `Config` aggregate of the layers' config structs. Each layer owns its config struct with `yaml` tags; the file `configs/config.yaml` holds the values, `configs/config.example.yaml` is the committed template.
- A mandatory parameter is checked on startup; an optional one has a default in `internal/config.Defaults()`. A parameter with neither is not allowed.
- A new parameter → a field in the layer's config + a line in `configs/config.example.yaml` in the existing section (`app`, `http`, `logger`, `postgres`, `llm`, `llm_reasoning`, `tokens`, `worker`, `kafka`, `auth`, `validator`).
- `cmd/*/main.go` — only `config.New` → `logger.Init` → `otel.Init` → `app.Run`. `log.Fatal` is allowed only here.

## 4. Transactions

- The only way to open a transaction is `transaction.Wrap(ctx, func(ctx context.Context) error { ... })`.
- Inside `Wrap` — repository method calls only. Kafka, HTTP, files, and any external side effects stay outside the transaction. An external call inside a transaction keeps it open for the duration of the network round-trip and is forbidden.
- The adapter gets its executor via `transaction.TryExtractTX(ctx)` — otherwise the call goes to the pool and ends up outside the transaction.
- A nested `transaction.Wrap` is forbidden.
- A multi-step write (several tables/operations that must be atomic together), including an outbox insert, must happen inside a single `Wrap`.
- `pkg/transaction` keeps `pool` via `Init()` and the `IsUnitTest` flag — this is a sanctioned exception to the ban on global state. It must not be extended with new globals; `IsUnitTest` may only be toggled in tests.

## 5. Errors

- Always wrap, with a prefix that is the called expression on that same line, without the receiver variable name:

```go
return fmt.Errorf("postgres.CreateProfile: %w", err)
return fmt.Errorf("json.Marshal: %w", err)
return fmt.Errorf("validate.Struct: %w", err)
return fmt.Errorf("pool.Begin: %w", err)
```

That is, `postgres.CreateProfile`, not `p.CreateProfile` and not `u.postgres.GetProfiles`.

- Domain errors are singletons; compare only with `errors.Is`, inspect the type with `errors.As`. Comparing `err.Error()` with a string is forbidden.
- At the boundary (controller) — `render.Error(ctx, err, "request failed")`.
- `pkg/render.Error` returns `fmt.Errorf("%s: %w", message, err)`, preserving the chain, and additionally puts the error into `ctxErr` from the context (`logger.ContextErrKey{}`), where `logger.Middleware` reads it from.
- Swallowing an error is forbidden. The only exception is a worker's background loop: `log.Error().Err(err)` and continuing the iteration. In usecase, adapter, and controller there is no "logged it and moved on".
- If an error is returned immediately, it needs no logging — the middleware or the worker will do that. Logging an error twice is noise.
- Forbidden: `panic` in business code, `log.Fatal` outside `cmd`, `os.Exit` outside `cmd`, returning an error upwards without wrapping.

## 6. Validation and domain mutation

- Entity invariants live in `domain`: constructor + `Validate()` + validator tags on the entity's fields (library per project, typically `go-playground/validator`).
- Input shape, pagination, sorting, and UUID format live in `dto.Validate()`.
- One constraint lives in exactly one place. Duplicating `required,min,max` across domain and dto is forbidden: the rule's home is chosen by meaning (is this an entity invariant or a request shape?).
- Order: input validation in usecase **before** touching the DB; invariant validation in the domain constructor/method.
- Domain mutation happens only through domain methods:

```go
// usecase
newProfile, err := profile.Apply(input) // inside domain: assignment + Validate()
```

Forbidden in usecase: `profile.Name = domain.Name(*input.Name)`, `profile.Contacts.Email = *input.Email`. Any assignment to domain entity fields outside the `domain` package is a violation.

## 7. Idempotency

- A mutating operation that publishes an event outwards must accept an `IdempotencyKey` field in the request body (the `IdempotencyKey` field in `dto.XxxInput`) and check it in the usecase.
- Mechanics in the usecase and the redis adapter:

1. `SetNX(key, marker, ttl)` — "am I the first?".
2. The key already exists → `Get(key)`: a finished result is found — return it and do not perform the operation; an in-flight marker is found — return `domain.ErrIdempotencyInProgress` (the boundary responds 409).
3. Perform the operation.
4. On success — `Set(key, <serialized dto Output>)` with the same TTL: a retry returns the same response with the same ID.

- The key is `idempotencyPrefix + <key>` and the TTL is a constant of the adapter package (`internal/adapter/redis`); the TTL value is the same for all operations.
- Only a `dto` Output may be stored, never a domain entity.
- `Get` + `Set` without atomicity is forbidden: between the check and the write two concurrent requests would both proceed.
- A marker without a stored result is not idempotency. A retry of a creating operation must return the same ID, not a nil `uuid.Nil`.

## 8. External side effects

- Kafka, HTTP calls, notifications — only after a successful transaction commit, never inside `Wrap` (§4).
- A side effect's error is returned upwards like any other. "Fire and forget" with a log is forbidden.
- If an event must not be lost, the outbox scheme is preferred: write the event into a table in the same transaction + a worker that reads it (`ReadOutboxKafka` + `worker.OutboxKafka`). A direct `Produce` is acceptable when the operation is idempotent (§7) and the client can safely retry.

## 9. Observability

- `tracer.Start` — in usecase and adapter, as the first line of the method, always with `defer span.End()`.
- Span name — `<layer> <tech> <Action>`: `usecase CreateProfile`, `adapter postgres GetProfile`, `adapter kafka Produce`, `transaction Wrap`.
- Spans are not opened in controllers or domain: transport is traced by middleware (`otel.Middleware`, the gRPC interceptor).
- Logs — the global logger, initialised at startup in `logger.Init` (typically `rs/zerolog`; per project — another one). There is no need to log what has already been returned as an error (§5). A meaningful log in an adapter exists only for events that have no channel to return an error (`Close`, a background loop).
- Metrics — `pkg/metrics`: `metrics.NewMiddleware` for HTTP, `metrics.NewProcess()`/`metrics.Total(...)` for an adapter's external operations. A new external call without a metric is an incomplete implementation.

## 10. Tests

- `domain` — table-driven unit tests next to the code (`profile_test.go`). Mandatory when invariants change.
- Integration — `test/integration/*_test.go` behind `//go:build integration`, a test suite (typically `testify/suite`), requests through the generated API client of this same service. A new endpoint → a new test file.
- Running them — only with commands from `Makefile`/`Justfile` (§0). The canonical targets: `test` = `go test -v -cover ./...`, `integration-test` = `go test -count=1 -v -tags=integration ./test/integration` (requires running infrastructure — e.g. `make up`). If the commands file lacks the target you need, ask instead of inventing a target name.
- Mocks — per project (if mockery is used: generated from `.mockery.yml` by the repository's generation command; never written by hand).
- A test is not fitted to the implementation: if the rule under test governs domain, the test calls domain, not usecase.

## 11. Naming, files, imports

- Files — `snake_case.go`, one file = one operation (`create_profile.go`, `get_profiles.go`). The package constructor and ports live in `<pkg>.go` (`postgres.go`, `usecase.go`, `v1.go`, `router.go`).
- Packages — lowercase, no `_`, no `utils`/`helpers`/`common`. A package is named after its layer or technology: `usecase`, `adapter/postgres`, `controller/http/v1`.
- A package may collide with stdlib (`http`, `grpc`, `redis`) — then the import must carry an alias: `http_server`, `pgpool`, `redislib`, `ver1`, `pb`.
- Aliases — only to resolve collisions and for generated code. An alias without a collision is noise.
- Port interfaces — named by role: `Postgres`, `Redis`, `Kafka`. Errors — `ErrXxx`. DTOs — `XxxInput`/`XxxOutput`.
- Imports: three groups in the order `standard` → third-party → local, as `gci` sections, applied automatically (`make lint` with `fix: true`).

## 12. Comments in code

Comments that restate what a function or a line does are not left behind: the code must read on its own.
If the code cannot be understood without an explanation, that is a reason to rewrite the code, not to add a comment.

A comment is left in only two cases:

1. **The agent violated a rule** of this file or the task contract (for example, bypassed §7 because the user explicitly asked). The text must name the violated rule and the reason.
2. **A stub or unfinished work** — a place that will be finished by a separate task.

Format — level and text separated by a colon; the level is uppercase and may only come from the list:

- `FIX: <text>` — broken code or an obvious defect that needs repair (including a violated rule, but the rule itself need not be named).
- `TODO: <text>` — unfinished work or a stub that will be completed.
- `HACK: <text>` — a temporary or ugly solution that needs replacing.
- `WARN: <text>` — a potential problem or trap the reader should know about.
- `PERF: <text>` — a place of potential inefficiency for future optimisation.
- `NOTE: <text>` — the rationale for a non-obvious decision that would otherwise be taken for a mistake.
- `TEST: <text>` — a pointer to a missing or incomplete test for this case.

```go
TODO: implement outbox reading in a worker, see §8
FIX: §7 (idempotency) violated: the key is set after the operation, rework it to SETNX
WARN: the DSN is built by concatenation; leak when logging — fix in a separate task
```

Rules:

- One comment — one level. Levels outside the list and unnamed explanations ("here we...", "this loop is needed to...") are forbidden.
- A comment must be brief.
- Comment text is in English unless the project block says otherwise.
- A comment goes above the line or at the end of the line, using ordinary Go syntax.
- A stub without `TODO`, a deviation from a rule without `FIX`/`WARN` or `NOTE` — a reason to send the change back for rework.
- Generated code (`gen/**`) is not commented.
- `//nolint` comments are governed by the linter rule (§14); they are not comments of this section.

## 13. Checklist: a new operation end-to-end

1. **domain** — is there an entity/method? Invariants, constructor, a mutation method with validation, singleton errors.
2. **dto** — `XxxInput` (+ `Validate()`, if there is a shape) and `XxxOutput` in a separate file.
3. **usecase** — a method with `tracer.Start`, input validation, idempotency (§7) if needed, domain, `transaction.Wrap` for multi-step writes, side effects after the commit.
4. **port** — if a new method of an external system is needed: an interface in `usecase.go`, `//go:generate mockery`, regenerate the mocks.
5. **adapter** — the port implementation: a file per operation, `const sql`, `TryExtractTX`, `ToDomain()`.
6. **controller** — request→dto→response mapping, `render.Error`, domain error mapping, `nil` as the second value.
7. **api** — the contract in `api/http/*.yaml` or `api/grpc/*.proto`, then the generation commands from `Makefile`/`Justfile` (e.g. `make grpc_gen` / `make generate`). Editing `gen/**` by hand is forbidden.
8. **migration** — create a migration with the command from `Makefile`/`Justfile` (e.g. `make migrate-create name=...`), write up and down, then apply it (`make migrate-up`). If such commands do not exist, suggest adding them to the commands file.
9. **config** — a new parameter: a field in the layer's config + a line in `configs/config.example.yaml`.
10. **app** — a new dependency: created in `app.go` and closed in reverse order.
11. **test** — a table-driven test for domain/invariants, an integration test for the endpoint (§10).
12. **check** — the lint command (§0), `go build ./...`, unit tests, integration ones if needed — all with commands from `Makefile`/`Justfile`.

## 14. Anti-patterns

Never do this:

1. `return Xxx400JSONResponse{...}, err` — the strict handler will return 500 and lose the response body. The response has been formed → `nil`.
2. `errors.Is(err, pgx.ErrNoRows)` after `Exec` — a dead check. `CommandTag.RowsAffected()` is what is needed.
3. Unwrapping an error down to the root (`errors.Unwrap` in a loop) before returning — the wrapper chain is lost and the context disappears from the log. Return `fmt.Errorf("%s: %w", message, err)`.
4. Assigning domain entity fields in usecase (e.g. a helper `update(profile, input)`) — mutation lives in domain.
5. `log.Error` + swallowing the error in usecase/adapter/controller.
6. One constraint duplicated in domain and in dto.
7. `fmt.Sprintf` to build SQL from client-supplied values.
8. Importing `internal/*` from `pkg/*`.
9. `context` in domain.
10. A new package in `pkg/` that knows about our domain or our service.
11. Global state outside the sanctioned exceptions: `pkg/logger`, `pkg/transaction` (`pool`, `IsUnitTest`), `domain.validate`, `dto.validate`.
12. Eyeballing validation in a controller instead of `dto.Validate()`/domain.
13. Logic in `cmd`/`config` beyond initialisation.
14. Editing generated code in `gen/**`.
15. `//nolint` without a linter name. Format: `//nolint:gochecknoglobals // sanctioned global: validate`.
16. Ignoring a returned error when constructing a dependency in `app.go`.
17. Invented build/lint/test commands when the user has not provided a `Makefile`/`Justfile` — ask (§0.1).
18. A comment that restates what the code does (what a function or a line does) — the code must read on its own (§12).
19. A stub or a deviation from the rules without a comment carrying a `FIX`/`TODO`/`WARN`/`NOTE` level (§12).

## 15. Deliberate trade-offs

These are accepted decisions, not bugs. Changing them silently is forbidden; if a task requires departing from one, say so explicitly.

- `transaction` keeps the pool globally (`Init`) and switches behaviour via `IsUnitTest`.
- Controllers depend on the concrete `*usecase.UseCase` rather than an interface: DIP is not honoured at this boundary, for the sake of simplicity.
- `usecase.Postgres` is a wide interface (outbox included); ISP is deliberately not honoured.
- `pkg/transaction` depends on `pkg/postgres`.
- The idempotency key is set **before** the operation with a TTL of 1 hour: a failed operation blocks a legitimate retry until the TTL expires. Removing the key is a manual operator action (`redis-cli DEL idempotency:<key>`), not code.
- `dto` is a single package for all layers (transport and usecase share the same structs).

## 16. Modern Go idioms (per the version in go.mod)

Source: **Modern Go Guidelines** — JetBrains' official guide (github.com/JetBrains/go-modern-guidelines; it can also ship as the `use-modern-go` skill/CLI). Below is a selected subset of the rules; the full list (54 rules) is in the guide's repository/CLI.

Application rules:

- Determine the Go version from `go.mod`/`go.work` and apply a rule only if the project version ≥ the stated one. Do not use a feature whose version is above the project's version.
- In new code, write the modern form straight away; existing code is not refactored without an explicit request.
- If a rule conflicts with the local style, failing to apply the rule is acceptable only when it does not compile, changes behaviour, or does not apply to the code; in a disputed case, ask.
- Do not extend the table with rules "just because": an addition is a deliberate decision by the owner.

| id                               | Go   | Rule                                                                                                        |
| -------------------------------- | ---- | ----------------------------------------------------------------------------------------------------------- |
| `time_since`                     | 1.0  | `time.Since(start)` instead of `time.Now().Sub(start)`                                                      |
| `time_until`                     | 1.8  | `time.Until(deadline)` instead of `deadline.Sub(time.Now())`                                                |
| `errors_is`                      | 1.13 | `errors.Is(err, target)` instead of `err == target` — accounts for wrapping                                 |
| `any`                            | 1.18 | `any` instead of `interface{}`                                                                              |
| `atomic_types`                   | 1.19 | typed `atomic.Bool`/`atomic.Int64`/`atomic.Pointer[T]` instead of untyped atomic functions                  |
| `context_cancel_cause`           | 1.20 | `context.WithCancelCause`/`context.Cause` when the cancellation must carry a reason                         |
| `errors_join`                    | 1.20 | `errors.Join` instead of manually combining errors (`errors.Is` still works for each)                       |
| `context_after_func`             | 1.21 | `context.AfterFunc` for work triggered by context cancellation                                              |
| `context_timeout_deadline_cause` | 1.21 | timeout/deadline contexts with a cause when the caller looks at the cancellation reason                     |
| `maps_clone`                     | 1.21 | `maps.Clone` instead of copying a map by hand                                                               |
| `maps_delete_func`               | 1.21 | `maps.DeleteFunc` — delete entries by predicate                                                             |
| `min_max`                        | 1.21 | built-in `min`/`max` instead of manual comparisons                                                          |
| `slices_clip`                    | 1.21 | `slices.Clip` — drop unused capacity                                                                        |
| `slices_clone`                   | 1.21 | `slices.Clone` for a copy of a slice                                                                        |
| `slices_compact`                 | 1.21 | `slices.Compact` — drop adjacent duplicates                                                                 |
| `slices_contains`                | 1.21 | `slices.Contains` instead of a manual search loop                                                           |
| `slices_index`                   | 1.21 | `slices.Index`, absent — `-1`                                                                               |
| `slices_index_func`              | 1.21 | `slices.IndexFunc` for a predicate search                                                                   |
| `slices_max_min`                 | 1.21 | `slices.Max`/`slices.Min` instead of manual loops                                                           |
| `slices_reverse`                 | 1.21 | `slices.Reverse` instead of a manual swap loop                                                              |
| `slices_sort`                    | 1.21 | `slices.Sort` for ordered values                                                                            |
| `slices_sort_func`               | 1.21 | `slices.SortFunc` + `cmp.Compare` instead of `sort.Slice`; for stable sorting — `slices.SortStableFunc`     |
| `json_omitzero`                  | 1.24 | `omitzero` on bool/numeric/struct/time fields; `omitempty` — for strings, slices, maps                      |
| `testing_t_context`              | 1.24 | `t.Context()` in tests; do not pass it into `t.Cleanup`                                                     |
| `sync_waitgroup_go`              | 1.25 | `wg.Go` instead of manual `go f()` + `wg.Add`/`wg.Done`                                                     |
| `errors_as_type`                 | 1.26 | `errors.AsType[T](err)` instead of a manual type assertion when checking a concrete error type              |
| `json_v2`                        | 1.27 | in new code — `encoding/json/v2` (stricter defaults); do not change existing `encoding/json` without an explicit request |
