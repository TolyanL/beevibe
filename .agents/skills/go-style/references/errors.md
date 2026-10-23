# Errors

Rules — AGENTS.md §5. Here — details and examples.

## Wrap prefix

The prefix is the called expression on the same line, without the receiver variable name:

| Write                                            | Don't write                               |
| ------------------------------------------------ | ----------------------------------------- |
| `fmt.Errorf("postgres.CreateProfile: %w", err)`  | `fmt.Errorf("p.CreateProfile: %w", err)`  |
| `fmt.Errorf("u.postgres.GetProfiles: %w", err)`  | `fmt.Errorf("get profiles: %w", err)`     |
| `fmt.Errorf("json.Marshal: %w", err)`            | `fmt.Errorf("marshal: %w", err)`          |
| `fmt.Errorf("validate.Struct: %w", err)`         | —                                         |
| `fmt.Errorf("pool.Begin: %w", err)`              | —                                         |
| `fmt.Errorf("tx.Commit: %w", err)`               | —                                         |
| `fmt.Errorf("domain.NewProfile: %w", err)`       | —                                         |

Wrap text is lowercase, no trailing period (Go convention).

## Domain errors — singletons

```go
var (
	ErrNotFound               = errors.New("not found")
	ErrUUIDInvalid            = errors.New("uuid is invalid")
	ErrIdempotencyInProgress  = errors.New("idempotency key already in progress")
	ErrIdempotencyKeyRequired = errors.New("idempotency key required")
)
```

- Only `errors.New`, no arguments, no dynamism (`err113`).
- Comparison — `errors.Is`; type inspection — `errors.As`. Never compare
  `err.Error()` against strings.
- Domain errors are not formatted and carry no request data — the data
  goes into the wraps.

## render.Error (controller boundary)

The canonical implementation preserves the chain (looping `errors.Unwrap` is forbidden,
it tears the context):

```go
func Error(ctx context.Context, err error, message string) error {
	if ctxErr, ok := ctx.Value(logger.ContextErrKey{}).(*error); ok {
		*ctxErr = fmt.Errorf("%s: %w", message, err)
	}

	return fmt.Errorf("%s: %w", message, err)
}
```

- Called once at the boundary, before forming the response.
- The controller's second return value is `nil` when a response has been formed
  (4xx included): returning `err` together with a response produces a 500 and the
  body is lost.

## Status codes

| Domain error                                       | HTTP | gRPC                    |
| --------------------------------------------------- | ---- | ----------------------- |
| `ErrNotFound`                                       | 404  | `codes.NotFound`        |
| `ErrIdempotencyInProgress`                          | 409  | `codes.AlreadyExists`   |
| `ErrUUIDInvalid`, the rest (incl. 400 validation)    | 400  | `codes.InvalidArgument` |
| unexpected                                           | 500  | `codes.Internal`        |

## Forbidden

- `log.Error` + swallowing the error in usecase/adapter/controller. Allowed
  only in a worker's background loop.
- Double logging: an error that was already returned is logged by the middleware itself.
- `panic` and `log.Fatal` in business code; an error returned upwards without a wrap.
