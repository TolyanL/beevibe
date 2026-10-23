package transaction

import (
	"context"
	"errors"
	"testing"
)

func TestWrapRunsFunctionInUnitMode(t *testing.T) {
	IsUnitTest = true

	called := false

	err := Wrap(context.Background(), func(ctx context.Context) error {
		called = true

		return nil
	})
	if err != nil {
		t.Fatalf("Wrap() error = %v, want nil", err)
	}

	if !called {
		t.Error("Wrap() did not call the function")
	}
}

func TestWrapPropagatesError(t *testing.T) {
	IsUnitTest = true

	cause := errors.New("write failed")

	err := Wrap(context.Background(), func(ctx context.Context) error {
		return cause
	})
	if !errors.Is(err, cause) {
		t.Fatalf("Wrap() error = %v, want %v", err, cause)
	}
}

func TestTryExtractTXFallsBackToPool(t *testing.T) {
	IsUnitTest = true

	if TryExtractTX(context.Background()) == nil {
		t.Error("TryExtractTX() = nil, want pool fallback")
	}
}
