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

func TestGetSessionUserIDReturnsOwner(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(userID, nil)

	got, err := uc.GetSessionUserID(context.Background(), "session-token")
	if err != nil {
		t.Fatalf("GetSessionUserID() error = %v, want nil", err)
	}

	if got != userID {
		t.Errorf("user id = %s, want %s", got, userID)
	}
}

func TestGetSessionUserIDRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	uc, _, _ := newTestUseCase(t)

	_, err := uc.GetSessionUserID(context.Background(), "")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("GetSessionUserID() error = %v, want ErrUnauthorized", err)
	}
}

func TestGetSessionUserIDRejectsUnknownToken(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	store.EXPECT().GetSessionUserID(mock.Anything, "unknown-token").Return(uuid.Nil, domain.ErrNotFound)

	_, err := uc.GetSessionUserID(context.Background(), "unknown-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("GetSessionUserID() error = %v, want ErrUnauthorized", err)
	}
}

func TestLogoutUserDeletesSession(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	store.EXPECT().DeleteSession(mock.Anything, "session-token").Return(nil)

	if _, err := uc.LogoutUser(context.Background(), dto.LogoutUserInput{SessionToken: "session-token"}); err != nil {
		t.Fatalf("LogoutUser() error = %v, want nil", err)
	}
}

func TestLogoutUserSkipsEmptyToken(t *testing.T) {
	t.Parallel()

	uc, _, _ := newTestUseCase(t)

	if _, err := uc.LogoutUser(context.Background(), dto.LogoutUserInput{}); err != nil {
		t.Fatalf("LogoutUser() error = %v, want nil", err)
	}
}

func TestTopupTokensCreditsBalance(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()
	user := domain.User{ID: userID, Email: domain.Email(testEmail), TokenBalance: 100}

	store.EXPECT().SumTopupsSince(mock.Anything, userID, mock.Anything).Return(int64(0), nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(user, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(updated domain.User) bool {
		return updated.TokenBalance == 100+250_000
	})).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 250_000 && op.Reason == domain.ReasonTopup
	})).Return(nil)

	output, err := uc.TopupTokens(context.Background(), dto.TopupTokensInput{UserID: userID, Amount: 250_000})
	if err != nil {
		t.Fatalf("TopupTokens() error = %v, want nil", err)
	}

	if output.Balance != 100+250_000 {
		t.Errorf("balance = %d, want %d", output.Balance, int64(100+250_000))
	}
}

func TestTopupTokensRejectsAmountOutsideLimits(t *testing.T) {
	t.Parallel()

	uc, _, _ := newTestUseCase(t)

	for _, amount := range []int64{0, 9_999, 1_000_001} {
		_, err := uc.TopupTokens(context.Background(), dto.TopupTokensInput{UserID: uuid.New(), Amount: amount})
		if err == nil {
			t.Fatalf("TopupTokens(%d) error = nil, want a limit error", amount)
		}
	}
}

func TestTopupTokensRejectsWhenDailyLimitReached(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().SumTopupsSince(mock.Anything, userID, mock.Anything).Return(int64(4_900_000), nil)

	_, err := uc.TopupTokens(context.Background(), dto.TopupTokensInput{UserID: userID, Amount: 200_000})
	if !errors.Is(err, domain.ErrTopupLimitReached) {
		t.Fatalf("TopupTokens() error = %v, want ErrTopupLimitReached", err)
	}
}

func TestGetProfileReturnsBalanceGamesAndOps(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		TokenBalance: 777,
	}, nil)
	store.EXPECT().ListUserGames(mock.Anything, userID, dto.ListUserGamesInput{Limit: profileGamesLimit}).Return([]dto.ProfileGame{
		{ID: gameID, Prompt: "змейка", Status: "ready", TokensCost: 320},
	}, 1, nil)
	store.EXPECT().ListTokenOps(mock.Anything, userID, dto.ListTokenOpsInput{Limit: profileGamesLimit}).Return([]domain.TokenOp{
		{UserID: userID, GameID: &gameID, Delta: -320, Reason: domain.ReasonGenerate},
	}, 1, nil)

	output, err := uc.GetProfile(context.Background(), dto.GetProfileInput{UserID: userID})
	if err != nil {
		t.Fatalf("GetProfile() error = %v, want nil", err)
	}

	if output.Balance != 777 {
		t.Errorf("balance = %d, want 777", output.Balance)
	}

	if len(output.Games) != 1 || output.Games[0].Prompt != "змейка" {
		t.Errorf("games = %+v, want one game", output.Games)
	}

	if len(output.TokenOps) != 1 || output.TokenOps[0].Reason != string(domain.ReasonGenerate) {
		t.Errorf("token ops = %+v, want one generate op", output.TokenOps)
	}

	if output.Prices.Generate != 320_000 || output.Prices.Patch != 40_000 {
		t.Errorf("prices = %+v, want generate 320000 patch 40000", output.Prices)
	}
}

func TestGetMeReturnsUser(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		TokenBalance: 42,
	}, nil)

	output, err := uc.GetMe(context.Background(), dto.GetMeInput{UserID: userID})
	if err != nil {
		t.Fatalf("GetMe() error = %v, want nil", err)
	}

	if output.User.Email != testEmail || output.User.Balance != 42 {
		t.Errorf("user = %+v, want email %s and balance 42", output.User, testEmail)
	}
}
