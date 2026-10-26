package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New(validator.WithRequiredStructEnabled()) //nolint:gochecknoglobals

type Email string

type PasswordHash string

const DisplayNameMaxLength = 40

type User struct {
	ID           uuid.UUID    `json:"id"`
	Email        Email        `json:"email"          validate:"required,email"`
	DisplayName  string       `json:"display_name"   validate:"max=40"`
	PasswordHash PasswordHash `json:"-"               validate:"required"`
	TokenBalance int64        `json:"token_balance"`
	CreatedAt    time.Time    `json:"created_at"`
}

func NewUser(email, passwordHash string, balance int64) (User, error) {
	if balance < 0 {
		return User{}, ErrInvalidInput
	}

	u := User{
		ID:           uuid.New(),
		Email:        Email(email),
		PasswordHash: PasswordHash(passwordHash),
		TokenBalance: balance,
		CreatedAt:    time.Now().UTC(),
	}

	if err := u.Validate(); err != nil {
		return User{}, fmt.Errorf("u.Validate: %w", err)
	}

	return u, nil
}

func (u *User) SetDisplayName(name string) error {
	name = strings.TrimSpace(name)

	if len([]rune(name)) > DisplayNameMaxLength {
		return fmt.Errorf("%w: display name is longer than %d characters", ErrInvalidInput, DisplayNameMaxLength)
	}

	if hasControlRunes(name) {
		return fmt.Errorf("%w: display name must not contain control characters", ErrInvalidInput)
	}

	u.DisplayName = name

	return nil
}

func (u User) PublicName() string {
	if name := strings.TrimSpace(u.DisplayName); name != "" {
		return name
	}

	local := string(u.Email)

	if at := strings.Index(local, "@"); at > 0 {
		local = local[:at]
	}

	return local
}

func (u User) Validate() error {
	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("validate.Struct: %w", err)
	}

	return nil
}

func (u *User) Debit(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("%w: debit amount must be positive", ErrInvalidInput)
	}

	if u.TokenBalance < amount {
		return ErrNotEnoughTokens
	}

	u.TokenBalance -= amount

	return nil
}

func (u *User) Credit(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("%w: credit amount must be positive", ErrInvalidInput)
	}

	u.TokenBalance += amount

	return nil
}
