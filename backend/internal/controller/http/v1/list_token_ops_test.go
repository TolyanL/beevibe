package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListTokenOpsReturnsPage(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().ListTokenOps(mock.Anything, userID, dto.ListTokenOpsInput{Limit: 10, Offset: 5}).
		Return([]domain.TokenOp{
			{UserID: userID, Delta: -320, Reason: domain.ReasonGenerate, CreatedAt: time.Now().UTC()},
		}, 30, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/me/token-ops?limit=10&offset=5", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.ListTokenOpsOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if len(output.Items) != 1 || output.Total != 30 || output.Items[0].Reason != "generate" {
		t.Errorf("output = %+v, want one generate op with total 30", output)
	}
}

func TestListTokenOpsRequiresAuth(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/me/token-ops", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
