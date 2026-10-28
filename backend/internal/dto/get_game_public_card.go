package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

type GetGamePublicCardInput struct {
	Slug     string
	ViewerID *uuid.UUID
}

type GetGamePublicCardOutput struct {
	ID          uuid.UUID         `json:"id"`
	Prompt      string            `json:"prompt"`
	Description string            `json:"description"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Author      string            `json:"author"`
	Cover       *domain.GameCover `json:"cover,omitempty"`
	Genre       string            `json:"genre"`
	CustomGenre string            `json:"custom_genre,omitempty"`
	CustomStyle string            `json:"custom_style,omitempty"`
	Style       string            `json:"style"`
	MapSize     string            `json:"map_size"`
	CreatedAt   time.Time         `json:"created_at"`
	Rating      RatingAggregate   `json:"rating"`
	Playtime    PlaytimeAggregate `json:"playtime"`
	MyRating    int               `json:"my_rating"`
	IsOwner     bool              `json:"is_owner"`
}
