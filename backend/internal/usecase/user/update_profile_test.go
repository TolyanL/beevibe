package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestUpdateProfileStoresDisplayName(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:    userID,
		Email: domain.Email(testEmail),
	}, nil)
	store.EXPECT().UpdateDisplayName(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.DisplayName == "Кристальный мастер"
	})).Return(nil)

	output, err := uc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
		UserID:      userID,
		DisplayName: "  Кристальный мастер ",
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v, want nil", err)
	}

	if output.User.DisplayName != "Кристальный мастер" {
		t.Errorf("display name = %q, want trimmed value", output.User.DisplayName)
	}
}

func TestUpdateProfileRejectsTooLongName(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	userID := uuid.New()

	store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:    userID,
		Email: domain.Email(testEmail),
	}, nil).Maybe()

	if _, err := uc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
		UserID:      userID,
		DisplayName: string(make([]rune, domain.DisplayNameMaxLength+1)),
	}); err == nil {
		t.Fatal("UpdateProfile() error = nil, want an error")
	}
}
