package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestLoginUserCreatesSession(t *testing.T) {
	t.Parallel()

	uc, store, auth := newTestUseCase(t)

	userID := uuid.New()
	user := domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		PasswordHash: domain.PasswordHash(testHash),
		TokenBalance: 500,
	}

	store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil)
	auth.EXPECT().CheckPassword(testHash, testPassword).Return(nil)
	auth.EXPECT().GenerateToken().Return("session-token", nil)
	store.EXPECT().CreateSession(mock.Anything, "session-token", user.ID, mock.Anything).Return(nil)

	output, err := uc.LoginUser(context.Background(), dto.LoginUserInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("LoginUser() error = %v, want nil", err)
	}

	if output.User.ID != user.ID {
		t.Errorf("user id = %s, want %s", output.User.ID, user.ID)
	}

	if output.SessionToken != "session-token" {
		t.Errorf("session token = %q, want %q", output.SessionToken, "session-token")
	}
}

func TestLoginUserRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	uc, store, auth := newTestUseCase(t)

	user := domain.User{ID: uuid.New(), Email: domain.Email(testEmail), PasswordHash: domain.PasswordHash(testHash)}

	store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil)
	auth.EXPECT().CheckPassword(testHash, "wrong-password").Return(errors.New("mismatch"))

	_, err := uc.LoginUser(context.Background(), dto.LoginUserInput{
		Email:    testEmail,
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("LoginUser() error = %v, want ErrUnauthorized", err)
	}
}

func TestLoginUserRejectsUnknownEmail(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(domain.User{}, domain.ErrNotFound)

	_, err := uc.LoginUser(context.Background(), dto.LoginUserInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("LoginUser() error = %v, want ErrUnauthorized", err)
	}
}

func TestLoginUserRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	uc, _, _ := newTestUseCase(t)

	_, err := uc.LoginUser(context.Background(), dto.LoginUserInput{
		Email:    "not-an-email",
		Password: "",
	})
	if err == nil {
		t.Fatal("LoginUser() error = nil, want validation error")
	}
}
