package game

import (
	"encoding/json"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

const (
	artifactPlan   = "plan"
	artifactCode   = "code"
	artifactAssets = "assets"
	artifactHTML   = "html"
)

// resumeStage возвращает этап, с которого нужно продолжить конвейер: следующим
// после последнего успешно завершённого.
func resumeStage(events []domain.GenEvent) domain.Status {
	lastDone := domain.Status("")

	for _, event := range events {
		if event.Type != domain.GenEventStage {
			continue
		}

		var stage dto.Stage

		if err := json.Unmarshal(event.Payload, &stage); err != nil {
			continue
		}

		if stage.Status != stageStatusDone {
			continue
		}

		if status := domain.Status(stage.Stage); status.IsPipeline() {
			lastDone = status
		}
	}

	return lastDone.Next()
}
