package catalog

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) List(ctx context.Context, input dto.ListCatalogInput) (dto.ListCatalogOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase catalog List")
	defer span.End()

	var output dto.ListCatalogOutput

	input.Trim()

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	cards, total, err := u.store.ListCatalog(ctx, dto.CatalogFilter{
		Q:      input.Q,
		Sort:   input.Sort,
		Genre:  input.Genre,
		Limit:  limit,
		Offset: input.Offset,
	})
	if err != nil {
		return output, fmt.Errorf("store.ListCatalog: %w", err)
	}

	output.Items = cards
	output.Total = total

	return output, nil
}
