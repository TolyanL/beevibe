package game

import (
	"encoding/json"

	"github.com/TolyanL/beevibe/internal/domain"
)

func specJSON(spec *domain.GameSpec) string {
	if spec == nil {
		return ""
	}

	raw, err := json.Marshal(spec)
	if err != nil {
		return ""
	}

	return string(raw)
}

func configJSON(cfg *domain.GameConfig) string {
	if cfg == nil {
		return ""
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		return ""
	}

	return string(raw)
}
