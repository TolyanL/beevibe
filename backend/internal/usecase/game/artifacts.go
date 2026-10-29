package game

import (
	"encoding/json"
	"errors"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

const (
	artifactReviewIssues = "review_issues"
	artifactPlan         = "plan"
	artifactCode         = "code"
	artifactAssets       = "assets"
	artifactHTML         = "html"
)

// NOTE: stops the pipeline without failing it: the author decides next.
var errNeedsDecision = errors.New("pipeline needs a decision")

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
