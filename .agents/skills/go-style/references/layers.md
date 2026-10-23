# Canonical layer examples

Working templates for each layer. Rules — in `AGENTS.md`; here — what the code
looks like. Follow the structure and naming exactly as shown.

## domain

```go
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Name string

type Age int

var (
	ErrNotFound = errors.New("not found")
	ErrNameTooShort = errors.New("name is too short")
)

var validate = validator.New(validator.WithRequiredStructEnabled()) //nolint:gochecknoglobals

type Profile struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"-"`
	Name      Name      `json:"name"     validate:"required,min=3,max=64"`
	Age       Age       `json:"age"      validate:"required,min=18,max=120"`
	Status    Status    `json:"status"`
	Verified  bool      `json:"verified"`
	Contacts  Contacts  `json:"contacts"`
}

// ProfileUpdate — partial update: a nil field is left unchanged, the result is
// validated in Apply. The domain keeps its own input type, dto does not import it.
type ProfileUpdate struct {
	Name  *string
	Age   *int
	Email *string
	Phone *string
}

func NewProfile(name string, age int, email, phone string) (Profile, error) {
	p := Profile{
		ID:       uuid.New(),
		Name:     Name(name),
		Age:      Age(age),
		Status:   Pending,
		Contacts: Contacts{Email: email, Phone: phone},
	}

	if err := p.Validate(); err != nil {
		return Profile{}, fmt.Errorf("p.Validate: %w", err)
	}

	return p, nil
}

func (p Profile) Validate() error {
	if err := validate.Struct(p); err != nil {
		return fmt.Errorf("validate.Struct: %w", err)
	}

	return nil
}

func (p Profile) Apply(u ProfileUpdate) (Profile, error) {
	if u.Name != nil {
		p.Name = Name(*u.Name)
	}

	if u.Age != nil {
		p.Age = Age(*u.Age)
	}

	if u.Email != nil {
		p.Contacts.Email = *u.Email
	}

	if u.Phone != nil {
		p.Contacts.Phone = *u.Phone
	}

	return p, p.Validate()
}

func (p Profile) IsDeleted() bool {
	return !p.DeletedAt.IsZero()
}

func (p Profile) ToEvent(topic string) (Event, error) { /* serialization into Event */ }
```

## dto

```go
package dto

type UpdateProfileInput struct {
	ID             string  `json:"id"`
	Name           *string `json:"name"`
	Age            *int    `json:"age"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	IdempotencyKey string  `json:"idempotency_key" validate:"required"`
}

func (i UpdateProfileInput) Validate() error {
	if err := validate.Struct(i); err != nil {
		return fmt.Errorf("validate.Struct: %w", err)
	}

	return nil
}
```

## usecase

```go
func (u *UseCase) CreateProfile(ctx context.Context, input dto.CreateProfileInput) (dto.CreateProfileOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase CreateProfile")
	defer span.End()

	var output dto.CreateProfileOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	profile, err := domain.NewProfile(input.Name, input.Age, input.Email, input.Phone)
	if err != nil {
		return output, fmt.Errorf("domain.NewProfile: %w", err)
	}

	event, err := profile.ToEvent("profile.created")
	if err != nil {
		return output, fmt.Errorf("profile.ToEvent: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.postgres.CreateProfile(ctx, profile); err != nil {
			return fmt.Errorf("postgres.CreateProfile: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	// Side effects — only after a successful commit (AGENTS.md §8).
	if err := u.kafka.Produce(ctx, event); err != nil {
		return output, fmt.Errorf("kafka.Produce: %w", err)
	}

	return dto.CreateProfileOutput{ID: profile.ID}, nil
}
```

The idempotent version (required for create/publish) — see
`references/idempotency.md`.

## adapter (postgres)

```go
package postgres

func (p *Postgres) GetProfile(ctx context.Context, profileID uuid.UUID) (domain.Profile, error) {
	ctx, span := tracer.Start(ctx, "adapter postgres GetProfile")
	defer span.End()

	const sql = `SELECT created_at, updated_at, deleted_at, name, age, status, verified, contacts
	             FROM profile WHERE id = $1`

	row := profileRow{}
	txOrPool := transaction.TryExtractTX(ctx)

	err := txOrPool.QueryRow(ctx, sql, profileID).Scan(row.Dest()...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Profile{}, domain.ErrNotFound
		}

		return domain.Profile{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row.ToDomain()
}

type profileRow struct {
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
	Name      pgtype.Text
	Age       pgtype.Int4
	Status    pgtype.Text
	Verified  pgtype.Bool
	Contacts  []byte
}

func (r *profileRow) Dest() []any {
	return []any{&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt, &r.Name, &r.Age, &r.Status, &r.Verified, &r.Contacts}
}

func (r *profileRow) ToDomain() (domain.Profile, error) { /* pgtype → domain, json.Unmarshal contacts */ }
```

UPDATE — with a `RowsAffected()` check, not `pgx.ErrNoRows`:

```go
func (p *Postgres) UpdateProfile(ctx context.Context, profile domain.Profile) error {
	const sql = `UPDATE profile SET name = $1, age = $2, contacts = $3, updated_at = NOW()
	             WHERE id = $4`

	tag, err := transaction.TryExtractTX(ctx).Exec(ctx, sql, profile.Name, profile.Age, contacts, profile.ID)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
```

## controller (http strict handler)

```go
func (h *Handlers) CreateProfile(ctx context.Context, request http_server.CreateProfileRequestObject,
) (http_server.CreateProfileResponseObject, error) {
	input := dto.CreateProfileInput{
		Name:  request.Body.Name,
		Age:   request.Body.Age,
		Email: request.Body.Email,
		Phone: request.Body.Phone,
	}

	output, err := h.usecase.CreateProfile(ctx, input)
	if err != nil {
		err = render.Error(ctx, err, "request failed")

		return http_server.CreateProfile400JSONResponse{Error: err.Error()}, nil
	}

	return http_server.CreateProfile200JSONResponse{ID: output.ID}, nil
}
```

Domain error mapping via `errors.Is`:

```go
switch {
case errors.Is(err, domain.ErrNotFound):
	return http_server.GetProfileByID404JSONResponse{Error: err.Error()}, nil
case errors.Is(err, domain.ErrIdempotencyInProgress):
	return http_server.UpdateProfile409JSONResponse{Error: err.Error()}, nil
default:
	return http_server.UpdateProfile400JSONResponse{Error: err.Error()}, nil
}
```

## app / config / cmd

```go
func Run(ctx context.Context, c config.Config) error {
	pgPool, err := pgpool.New(ctx, c.Postgres)
	if err != nil {
		return fmt.Errorf("pgpool.New: %w", err)
	}

	transaction.Init(pgPool)

	uc := usecase.New(postgres.New(), redis.New(redisClient), kafkaProducer)

	// ... controllers, workers

	<-sig // shutdown: controllers → adapters in reverse order
	return nil
}
```

```go
type App struct {
	Name    string `envconfig:"APP_NAME"    required:"true"`
	Version string `envconfig:"APP_VERSION" required:"true"`
}

type Config struct {
	Port string `default:"8080" envconfig:"HTTP_PORT"`
}
```

```go
func main() {
	c, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("config.New")
	}

	logger.Init(c.Logger)
	ctx := context.Background()

	if err := otel.Init(ctx, c.OTEL); err != nil {
		log.Fatal().Err(err).Msg("otel.Init")
	}
	defer otel.Close()

	if err := app.Run(ctx, c); err != nil {
		log.Error().Err(err).Msg("app.Run")
	}
}
```
