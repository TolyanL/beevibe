package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListTokenOpsPagesAndDefaults(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name   string
		input  dto.ListTokenOpsInput
		want   dto.ListTokenOpsInput
		wantOp []domain.TokenOp
		total  int
	}{
		{
			name:  "default limit",
			input: dto.ListTokenOpsInput{},
			want:  dto.ListTokenOpsInput{Limit: 20},
			wantOp: []domain.TokenOp{
				{UserID: userID, Delta: -320, Reason: domain.ReasonGenerate},
			},
			total: 1,
		},
		{
			name:   "explicit paging",
			input:  dto.ListTokenOpsInput{Limit: 5, Offset: 10},
			want:   dto.ListTokenOpsInput{Limit: 5, Offset: 10},
			wantOp: []domain.TokenOp{},
			total:  0,
		},
	}

	for _, tt := range tests {
		uc, store, _ := newTestUseCase(t)

		store.EXPECT().ListTokenOps(mock.Anything, userID, tt.want).Return(tt.wantOp, tt.total, nil)

		output, err := uc.ListTokenOps(context.Background(), userID, tt.input)
		if err != nil {
			t.Fatalf("%s: ListTokenOps() error = %v, want nil", tt.name, err)
		}

		if output.Total != tt.total || len(output.Items) != len(tt.wantOp) {
			t.Errorf("%s: output = %+v, want total %d and %d items", tt.name, output, tt.total, len(tt.wantOp))
		}
	}
}

func TestListTokenOpsRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	uc, _, _ := newTestUseCase(t)

	if _, err := uc.ListTokenOps(context.Background(), userID, dto.ListTokenOpsInput{Limit: 1000}); err == nil {
		t.Fatal("ListTokenOps() error = nil, want validation error")
	}
}
