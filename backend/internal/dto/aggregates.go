package dto

import (
	"time"

	"github.com/google/uuid"
)

type RatingAggregate struct {
	Cool        int     `json:"cool"`
	Bad         int     `json:"bad"`
	Total       int     `json:"total"`
	CoolPercent float64 `json:"cool_percent"`
}

type PlaytimeAggregate struct {
	Sessions int     `json:"sessions"`
	Seconds  int     `json:"seconds"`
	Hours    float64 `json:"hours"`
}

type CatalogFilter struct {
	Sort   string
	Genre  string
	Limit  int
	Offset int
}

type CatalogCard struct {
	ID         uuid.UUID         `json:"id"`
	Prompt     string            `json:"prompt"`
	Genre      string            `json:"genre"`
	MapSize    string            `json:"map_size"`
	Style      string            `json:"style"`
	CreatedAt  time.Time         `json:"created_at"`
	Rating     RatingAggregate   `json:"rating"`
	Playtime   PlaytimeAggregate `json:"playtime"`
	TokensCost int               `json:"tokens_cost"`
	ModelCalls int               `json:"model_calls"`
}
