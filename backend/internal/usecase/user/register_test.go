package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	usermocks "github.com/TolyanL/beevibe/internal/usecase/mocks/user"
)

const (
	testEmail    = "user@example.com"
	testPassword = "password123"
	testHash     = "hashed-password"
)

func newTestUseCase(t *testing.T) (*UseCase, *usermocks.MockStore, *usermocks.MockAuth) {
	t.Helper()

	store := usermocks.NewMockStore(t)
	auth := usermocks.NewMockAuth(t)

	uc := New(store, auth, Tokens{
		GrantSignup:   1_000_000,
		TopupMin:      10_000,
		TopupMax:      1_000_000,
		TopupDailyCap: 5_000_000,
		DemoTopup:     true,
		PriceGenerate: 320_000,
		PricePatch:    40_000,
		PriceMeta:     10_000,
	})

	return uc, store, auth
}

func TestRegisterUserCreatesUserWithGrantAndSession(t *testing.T) {
	t.Parallel()

	uc, store, auth := newTestUseCase(t)

	auth.EXPECT().HashPassword(testPassword).Return(testHash, nil)
	auth.EXPECT().GenerateToken().Return("session-token", nil)

	store.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.Email == domain.Email(testEmail) &&
			user.PasswordHash == domain.PasswordHash(testHash) &&
			user.TokenBalance == 1_000_000
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 1_000_000 && op.Reason == domain.ReasonSignupGrant
	})).Return(nil)
	store.EXPECT().CreateSession(mock.Anything, "session-token", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.RegisterUser(context.Background(), dto.RegisterUserInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("RegisterUser() error = %v, want nil", err)
	}

	if output.User.Email != testEmail {
		t.Errorf("email = %q, want %q", output.User.Email, testEmail)
	}

	if output.User.Balance != 1_000_000 {
		t.Errorf("balance = %d, want %d", output.User.Balance, int64(1_000_000))
	}

	if output.SessionToken != "session-token" {
		t.Errorf("session token = %q, want %q", output.SessionToken, "session-token")
	}

	if output.User.ID == uuid.Nil {
		t.Error("user id = nil, want uuid")
	}
}

func TestRegisterUserRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	uc, _, _ := newTestUseCase(t)

	_, err := uc.RegisterUser(context.Background(), dto.RegisterUserInput{
		Email:    "not-an-email",
		Password: testPassword,
	})
	if err == nil {
		t.Fatal("RegisterUser() error = nil, want validation error")
	}
}

func TestRegisterUserFailsWhenStoreRejects(t *testing.T) {
	t.Parallel()

	uc, store, auth := newTestUseCase(t)

	auth.EXPECT().HashPassword(testPassword).Return(testHash, nil)
	auth.EXPECT().GenerateToken().Return("session-token", nil)
	store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(domain.ErrAlreadyExists)

	_, err := uc.RegisterUser(context.Background(), dto.RegisterUserInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("RegisterUser() error = %v, want ErrAlreadyExists", err)
	}
}

func TestRegisterUserFailsWhenHashFails(t *testing.T) {
	t.Parallel()

	uc, _, auth := newTestUseCase(t)

	auth.EXPECT().HashPassword(testPassword).Return("", errors.New("bcrypt failed"))

	_, err := uc.RegisterUser(context.Background(), dto.RegisterUserInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err == nil {
		t.Fatal("RegisterUser() error = nil, want error")
	}
}
