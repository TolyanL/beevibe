package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type RatingAggregate struct {
	Average float64 `json:"average"`
	Total   int     `json:"total"`
	Counts  [5]int  `json:"counts"`
}

type PlaytimeAggregate struct {
	Sessions int     `json:"sessions"`
	Seconds  int     `json:"seconds"`
	Hours    float64 `json:"hours"`
}

type CatalogFilter struct {
	Q      string
	Sort   string
	Genre  string
	Limit  int
	Offset int
}

type CatalogCard struct {
	ID          uuid.UUID         `json:"id"`
	Prompt      string            `json:"prompt"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Cover       *domain.GameCover `json:"cover,omitempty"`
	Author      string            `json:"author"`
	Slug        string            `json:"slug"`
	Genre       string            `json:"genre"`
	CustomGenre string            `json:"custom_genre,omitempty"`
	CustomStyle string            `json:"custom_style,omitempty"`
	MapSize    string            `json:"map_size"`
	Style      string            `json:"style"`
	CreatedAt  time.Time         `json:"created_at"`
	Rating     RatingAggregate   `json:"rating"`
	Playtime   PlaytimeAggregate `json:"playtime"`
	TokensCost int               `json:"tokens_cost"`
	ModelCalls int               `json:"model_calls"`
}
