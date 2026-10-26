package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/dto"
	catalogmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/catalog"
)

func newTestUseCase(t *testing.T) (*UseCase, *catalogmocks.MockStore) {
	t.Helper()

	store := catalogmocks.NewMockStore(t)

	return New(store), store
}

func TestListAppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	cards := []dto.CatalogCard{{ID: uuid.New(), Prompt: "змейка", Genre: "snake"}}

	store.EXPECT().ListCatalog(mock.Anything, dto.CatalogFilter{
		Sort:   "rating",
		Genre:  "snake",
		Limit:  20,
		Offset: 0,
	}).Return(cards, 40, nil)

	output, err := uc.List(context.Background(), dto.ListCatalogInput{
		Sort:  "rating",
		Genre: "snake",
	})
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}

	if len(output.Items) != 1 || output.Items[0].Prompt != "змейка" {
		t.Errorf("items = %+v, want one card", output.Items)
	}
}

func TestListKeepsExplicitLimit(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	store.EXPECT().ListCatalog(mock.Anything, dto.CatalogFilter{
		Limit:  5,
		Offset: 10,
	}).Return([]dto.CatalogCard{}, 0, nil)

	if _, err := uc.List(context.Background(), dto.ListCatalogInput{Limit: 5, Offset: 10}); err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
}

func TestListRejectsInvalidSort(t *testing.T) {
	t.Parallel()

	uc, _ := newTestUseCase(t)

	_, err := uc.List(context.Background(), dto.ListCatalogInput{Sort: "unknown"})
	if err == nil {
		t.Fatal("List() error = nil, want validation error")
	}
}

func TestListFailsWhenStoreFails(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	store.EXPECT().ListCatalog(mock.Anything, mock.Anything).Return(nil, 0, errors.New("db is down"))

	_, err := uc.List(context.Background(), dto.ListCatalogInput{})
	if err == nil {
		t.Fatal("List() error = nil, want error")
	}
}

func TestListSearchesAcrossWholeCatalog(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	store.EXPECT().ListCatalog(mock.Anything, dto.CatalogFilter{Q: "змей", Limit: 20, Offset: 0}).
		Return([]dto.CatalogCard{{ID: uuid.New(), Prompt: "неоновая змейка", Genre: "snake"}}, 2, nil)

	output, err := uc.List(context.Background(), dto.ListCatalogInput{Q: "  змей  "})
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}

	if len(output.Items) != 1 || output.Total != 2 {
		t.Errorf("output = %+v, want one item with total 2", output)
	}
}
