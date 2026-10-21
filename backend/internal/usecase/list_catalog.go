package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) ListCatalog(ctx context.Context, input dto.ListCatalogInput) (dto.ListCatalogOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase ListCatalog")
	defer span.End()

	var output dto.ListCatalogOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	cards, err := u.store.ListCatalog(ctx, dto.CatalogFilter{
		Sort:   input.Sort,
		Genre:  input.Genre,
		Limit:  limit,
		Offset: input.Offset,
	})
	if err != nil {
		return output, fmt.Errorf("store.ListCatalog: %w", err)
	}

	output.Items = cards

	return output, nil
}
