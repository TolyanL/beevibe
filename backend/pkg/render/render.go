package render

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/pkg/logger"
)

func Error(ctx context.Context, err error, message string) error {
	wrapped := fmt.Errorf("%s: %w", message, err)

	if ctxErr, ok := ctx.Value(logger.ContextErrKey{}).(*error); ok {
		*ctxErr = wrapped
	}

	return wrapped
}
