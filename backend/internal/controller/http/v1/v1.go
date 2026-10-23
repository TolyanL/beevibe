package v1

import (
	"github.com/TolyanL/beevibe/internal/usecase"
)

type Handlers struct {
	usecase *usecase.UseCase
}

func New(uc *usecase.UseCase) *Handlers {
	return &Handlers{usecase: uc}
}
