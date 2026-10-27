package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestResolveGameIDAcceptsSlug(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "neonovaya-zmeyka-1a2b").Return(domain.Game{ID: gameID}, nil)

	got, err := uc.ResolveGameID(context.Background(), "neonovaya-zmeyka-1a2b")
	if err != nil {
		t.Fatalf("ResolveGameID() error = %v, want nil", err)
	}

	if got != gameID {
		t.Errorf("id = %s, want %s", got, gameID)
	}
}

func TestResolveGameIDRejectsUUID(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	raw := uuid.New().String()

	// uuid больше не адрес игры: он ищется только как slug и не находится.
	store.EXPECT().GetGameBySlug(mock.Anything, raw).Return(domain.Game{}, domain.ErrNotFound)

	if _, err := uc.ResolveGameID(context.Background(), raw); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ResolveGameID() error = %v, want ErrNotFound", err)
	}
}

func TestResolveGameIDRejectsEmpty(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	if _, err := uc.ResolveGameID(context.Background(), ""); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("ResolveGameID() error = %v, want ErrInvalidInput", err)
	}
}
