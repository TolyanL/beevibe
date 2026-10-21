package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) GetSessionUserID(ctx context.Context, token string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "usecase GetSessionUserID")
	defer span.End()

	if token == "" {
		return uuid.Nil, domain.ErrUnauthorized
	}

	userID, err := u.store.GetSessionUserID(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return uuid.Nil, domain.ErrUnauthorized
		}

		return uuid.Nil, fmt.Errorf("store.GetSessionUserID: %w", err)
	}

	return userID, nil
}
