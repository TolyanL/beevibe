# Idempotency and external side effects

Rules — AGENTS.md §7, §8. Here — the mechanics in code.

## When idempotency is mandatory

A mutating operation that publishes an event outwards (create + kafka/HTTP) must
accept an `IdempotencyKey` and check it in the usecase. Pure reads — no.

The field lives in the request body (`dto.XxxInput`), not in a header:

```go
type CreateProfileInput struct {
	Name           string `json:"name"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}
```

## Mechanics (SETNX → result in Redis)

1. `SetNX(key, marker, ttl)` — "am I the first?". Only atomically; `Get` + `Set`
   separately are forbidden (race).
2. The key already exists → `Get(key)`:
   - a finished result → return it, do not perform the operation;
   - an in-flight marker → `domain.ErrIdempotencyInProgress` (boundary → 409).
3. Perform the operation.
4. Success → `Set(key, <serialized dto Output>)` with the same TTL: a retry
   returns the same ID.

The key and TTL are constants of the `internal/adapter/redis` package, the TTL is uniform:

```go
const (
	idempotencyPrefix = "idempotency:"
	ttl               = time.Hour
)
```

Only a `dto` Output may be stored, never a domain entity. A marker without a
stored result does not count as idempotency: a create retry must return the same
ID, not `uuid.Nil`.

## Redis adapter signatures

```go
func (r *Redis) IsExists(ctx context.Context, key string) bool   // marker/result already exists
func (r *Redis) SetNX(ctx context.Context, key string, marker []byte, ttl time.Duration) (bool, error)
func (r *Redis) SetResult(ctx context.Context, key string, output []byte, ttl time.Duration) error
```

## Scheme in usecase (create)

```go
key := input.IdempotencyKey // the key is already validated in dto

first, err := u.redis.SetNX(ctx, key, marker, ttl)
if err != nil {
	return output, fmt.Errorf("redis.SetNX: %w", err)
}

if !first {
	if u.redis.IsExists(ctx, key) {
		// this holds the result (step 4) OR an in-flight marker:
		// distinguish by the value contents; in-flight → ErrIdempotencyInProgress
		return u.storedResult(ctx, key)
	}

	return output, domain.ErrIdempotencyInProgress
}

// ... operation (transaction.Wrap) ...

// success:
result, _ := json.Marshal(dto.CreateProfileOutput{ID: profile.ID})
if err := u.redis.SetResult(ctx, key, result, ttl); err != nil {
	return output, fmt.Errorf("redis.SetResult: %w", err)
}

return dto.CreateProfileOutput{ID: profile.ID}, nil
```

## External side effects

- Kafka/HTTP/notifications — only after a successful commit, never inside
  `transaction.Wrap` (it keeps the transaction open for the network round-trip).
- A side effect's error is returned upwards; "sent and forgot" is forbidden.
- If an event must not be lost — outbox: a row in a table within the same transaction
  + a worker (`ReadOutboxKafka` + `worker.OutboxKafka`). A direct `Produce`
  is acceptable only when the operation is idempotent.
