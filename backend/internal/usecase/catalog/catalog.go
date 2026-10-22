package catalog

import (
	"context"

	"github.com/TolyanL/beevibe/internal/dto"
)

// Store — порт доступа к каталогу игр.
type Store interface {
	ListCatalog(ctx context.Context, filter dto.CatalogFilter) ([]dto.CatalogCard, error)
}

type UseCase struct {
	store Store
}

func New(store Store) *UseCase {
	return &UseCase{store: store}
}
