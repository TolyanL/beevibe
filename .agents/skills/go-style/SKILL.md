---
name: go-style
description: |
  House style for Go code: canonical examples of the domain/dto/usecase/adapter/controller
  layers, naming conventions, error wrapping, comments, tests. Load when writing,
  refactoring, or reviewing Go code in a service with clean architecture
  (internal/app, usecase, adapter, transaction.Wrap, AGENTS.md Go service).
---

# go-style — house style for Go code

Architecture and style rules live in the repo's `AGENTS.md`; details and code live here.
Priority: `AGENTS.md` (rules) > this skill (examples) > the linter (final arbiter).
A mismatch between the linter and the rules is fixed by editing the rules, not with `//nolint`.

This skill complements the generic `golang-patterns`: it pins the personal conventions,
not generally accepted practices.

The library stack and Go version are not fixed here: examples in `references/` are given
on a typical stack (pgx, chi, validator, zerolog, mockery...), take the actual stack
and Go version from the repository (`go.mod`, imports, the commands file) — the rules
apply with any stack.

## When to load

- writing or refactoring Go code in a service with an `AGENTS.md` Go service block;
- reviewing Go code;
- bootstrapping a new Go service (skeleton per AGENTS.md §1).

## What's here

| File                        | Contents                                                             |
| --------------------------- | -------------------------------------------------------------------- |
| `references/layers.md`      | canonical code examples for each layer (working templates)           |
| `references/errors.md`      | error wrapping, domain singletons, `render.Error`, status codes      |
| `references/idempotency.md` | idempotency (`SETNX`) and external side effects                      |
| `references/style.md`       | naming, file structure, imports, comments                            |
| `references/testing.md`     | domain unit tests and integration tests                              |

Read only the topic you need from `references/`, not the whole directory at once.

## Quick style rules

- Code reads without comments. A comment carries a level only:
  `FIX`/`TODO`/`HACK`/`WARN`/`PERF`/`NOTE`/`TEST` (AGENTS.md §12).
- File = one operation: `create_profile.go`, `get_profiles.go`; the package
  constructor goes in `<pkg>.go` (`postgres.go`, `usecase.go`, `v1.go`).
- Errors are always wrapped with a prefix that is the called expression:
  `fmt.Errorf("postgres.CreateProfile: %w", err)`.
- Domain: typed primitives (`type Name string`), constructors, mutation only
  through domain methods.
- Validation: invariants live in domain, input shape lives in `dto.Validate()`;
  one constraint lives in exactly one place.
- Transactions only via `transaction.Wrap`; DB access from an adapter only via
  `transaction.TryExtractTX(ctx)`.
- Everything an agent writes must pass the lint command from the repository's
  commands file (AGENTS.md §0).
- Modern Go idioms follow the version in `go.mod`: the rule list is in AGENTS.md §16
  (source — JetBrains Modern Go Guidelines); features above the project's version are
  not used.

## Go code review: layer-by-layer checklist

- **domain** — invariants in constructors/methods; error singletons; mutation
  only through methods; no `context`, `dto`, DB drivers.
- **dto** — `json` tags; `Validate()` only for input shape; no logic
  duplicating domain.
- **usecase** — `tracer.Start` + `defer span.End()`; input validation before the DB;
  multi-step writes in a single `Wrap`; errors not swallowed; no SQL.
- **adapter** — `TryExtractTX` only here; `Dest()`/`ToDomain()` for `Scan`;
  `RowsAffected()` after `Exec`; driver "no rows" → domain error.
- **controller** — request→dto→response mapping; `render.Error`; second
  return value `nil` when a response has been formed; mapping via
  `errors.Is`.
- **config/app/cmd** — required/default for every parameter; a line in
  `.env.example`; dependencies shut down in reverse order.
- **general** — lint command green; comments carry a level only; `gen/**` never
  edited by hand.
