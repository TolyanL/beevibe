package catalog

import (
	"context"

	"github.com/TolyanL/beevibe/internal/dto"
)

type Store interface {
	ListCatalog(ctx context.Context, filter dto.CatalogFilter) ([]dto.CatalogCard, int, error)
}

type UseCase struct {
	store Store
}

func New(store Store) *UseCase {
	return &UseCase{store: store}
}
