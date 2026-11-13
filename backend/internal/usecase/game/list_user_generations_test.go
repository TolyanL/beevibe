package game

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListUserGenerationsAppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()
	createdAt := time.Now().UTC()

	store.EXPECT().ListUserGenerations(mock.Anything, userID, dto.ListUserGenerationsInput{
		Limit:  20,
		Offset: 0,
	}).Return([]dto.ProfileGeneration{
		{GameID: uuid.New(), GameSlug: "zmeyka-1a2b", VersionNumber: 3, Source: "patch", CreatedAt: createdAt},
	}, 7, nil)

	output, err := uc.ListUserGenerations(context.Background(), dto.ListUserGenerationsInput{UserID: userID})
	if err != nil {
		t.Fatalf("ListUserGenerations() error = %v, want nil", err)
	}

	if len(output.Items) != 1 || output.Total != 7 {
		t.Errorf("output = %+v, want one generation with total 7", output)
	}
}

func TestListUserGenerationsKeepsExplicitPaging(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().ListUserGenerations(mock.Anything, userID, dto.ListUserGenerationsInput{
		Limit:  5,
		Offset: 10,
	}).Return([]dto.ProfileGeneration{}, 0, nil)

	output, err := uc.ListUserGenerations(context.Background(), dto.ListUserGenerationsInput{
		UserID: userID,
		Limit:  5,
		Offset: 10,
	})
	if err != nil {
		t.Fatalf("ListUserGenerations() error = %v, want nil", err)
	}

	if output.Total != 0 || len(output.Items) != 0 {
		t.Errorf("output = %+v, want an empty page", output)
	}
}

func TestListUserGenerationsRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	uc, _, _, _, _ := newTestUseCase(t)

	if _, err := uc.ListUserGenerations(context.Background(), dto.ListUserGenerationsInput{
		UserID: uuid.New(),
		Limit:  1000,
	}); err == nil {
		t.Fatal("ListUserGenerations() error = nil, want validation error")
	}

	if _, err := uc.ListUserGenerations(context.Background(), dto.ListUserGenerationsInput{
		UserID: uuid.New(),
		Offset: -1,
	}); err == nil {
		t.Fatal("ListUserGenerations() error = nil, want validation error")
	}
}

func TestListUserGenerationsFailsWhenStoreFails(t *testing.T) {
	t.Parallel()

	uc, store, _, _, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().ListUserGenerations(mock.Anything, userID, dto.ListUserGenerationsInput{Limit: 20}).
		Return(nil, 0, errors.New("db is down"))

	if _, err := uc.ListUserGenerations(context.Background(), dto.ListUserGenerationsInput{UserID: userID}); err == nil {
		t.Fatal("ListUserGenerations() error = nil, want error")
	}
}
