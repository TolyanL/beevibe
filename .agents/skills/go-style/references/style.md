# Style: naming, file structure, imports, comments

Rules — AGENTS.md §11, §12; language details here.

## Naming

- Local variables — short, meaningful: `err`, `cfg`, `row`, `out`.
  No transliteration and no Hungarian notation.
- Receiver — one or two letters of the type name: `u` (usecase), `p` (postgres/
  profile), `h` (handlers), `r` (redis/router/row), `s` (server), `w` (worker).
- Port method — VerbEntity → the same name as in the usecase interface:
  `CreateProfile`, `GetProfiles`, `ReadOutboxKafka`.
- Span — `<layer> <Action>`: `usecase CreateProfile`, `adapter postgres
  GetProfile`, `adapter kafka Produce`, `transaction Wrap`.
- File = one operation (`create_profile.go`); constructor and ports in `<pkg>.go`;
  the layer config lives in `<tech>_config.go` next to the package or in `<pkg>.go`.
- Errors — `Err<Name>`. DTO — `<Entity><Action>Input/Output`.

## File structure

Order top to bottom:

```go
package name            // required

import (                // groups: standard / third-party / local (gci)
	...
)

const ( ... )           // if present — first

var ( ... )             // error singletons, legalized globals

type Xxx struct { ... } // types

func NewXxx(...) ...    // constructor

func (x Xxx) Method(...) ...
```

- `const sql = ...` in an adapter — inside the operation file (for a single one) or in
  `<pkg>.go` (for several), as in the project's reference.
- The pair `ctx, span := tracer.Start(...)` + `defer span.End()` — as the first
  line of a public usecase/adapter method body.

## Comments

- Code should read on its own; restating what a function/line does is forbidden.
- A comment carries a level only: `FIX`/`TODO`/`HACK`/`WARN`/`PERF`/`NOTE`/
  `TEST` with text after the colon (AGENTS.md §12). One comment — one
  level.
- Where it's needed: the agent broke a rule (the text names the rule and the reason),
  a stub, a non-obvious decision, a notable trap.
- Comment text is in English unless the project block of AGENTS.md says otherwise.
- Separator comments inside a body are allowed as a scenario "table of contents"
  (inside a method — per the AGENTS.md §3.3 template), without restating the code.
- `//nolint:<linter>` — only with the concrete linter and a reason:
  `//nolint:gochecknoglobals // sanctioned global: validate`.

## Imports and aliases

- Three groups (gci): standard → default (third-party) → localmodule; the order is
  what the linter does with `fix: true` — don't rearrange by hand afterwards.
- An alias only for a stdlib collision or for generated code:
  `http_server` (oapi), `pb` (grpc), `pgpool` (pkg/postgres), `redislib`
  (pkg/redis), `ver1` (controller v1). An alias without a collision is noise.

## Prohibitions and obligations

- No `utils`/`helpers`/`common` packages; a package is named after its layer/technology.
- No `context` in domain; no importing `internal/*` from `pkg/*`.
- No SQL and no HTTP statuses in usecase.
- Every external adapter operation carries a metric (AGENTS.md §9) and a span.
