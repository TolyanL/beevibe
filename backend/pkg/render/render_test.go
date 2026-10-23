package render

import (
	"context"
	"errors"
	"testing"

	"github.com/TolyanL/beevibe/pkg/logger"
)

func TestErrorWrapsMessage(t *testing.T) {
	t.Parallel()

	cause := errors.New("db is down")

	err := Error(context.Background(), cause, "request failed")
	if err == nil {
		t.Fatal("Error() = nil, want error")
	}

	if err.Error() != "request failed: db is down" {
		t.Errorf("message = %q, want %q", err.Error(), "request failed: db is down")
	}

	if !errors.Is(err, cause) {
		t.Errorf("errors.Is() = false, want true for the cause")
	}
}

func TestErrorStoresCauseInContext(t *testing.T) {
	t.Parallel()

	cause := errors.New("db is down")
	ctxErr := new(error)
	ctx := context.WithValue(context.Background(), logger.ContextErrKey{}, ctxErr)

	err := Error(ctx, cause, "request failed")
	if !errors.Is(*ctxErr, cause) {
		t.Errorf("context error = %v, want wrapped cause", *ctxErr)
	}

	if err == nil {
		t.Fatal("Error() = nil, want error")
	}
}
