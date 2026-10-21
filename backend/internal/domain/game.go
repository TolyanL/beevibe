package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Genre string

const (
	GenreArcade     Genre = "arcade"
	GenreSnake      Genre = "snake"
	GenreClicker    Genre = "clicker"
	GenrePlatformer Genre = "platformer"
	GenrePuzzle     Genre = "puzzle"
	GenreCustom     Genre = "custom"
)

type MapSize string

const (
	MapSizeSmall  MapSize = "small"
	MapSizeMedium MapSize = "medium"
	MapSizeLarge  MapSize = "large"
)

type Style string

const (
	StylePixel   Style = "pixel"
	StyleCartoon Style = "cartoon"
	StyleNeon    Style = "neon"
	StyleMinimal Style = "minimal"
)

type Tempo int

const (
	TempoSlow Tempo = 1
	TempoFast Tempo = 10
)

type GameParams struct {
	Genre   Genre   `json:"genre,omitempty"`
	Tempo   Tempo   `json:"tempo,omitempty"`
	MapSize MapSize `json:"map_size,omitempty"`
	Style   Style   `json:"style,omitempty"`
}

type Game struct {
	ID         uuid.UUID   `json:"id"`
	OwnerID    uuid.UUID   `json:"owner_id"`
	Prompt     string      `json:"prompt"`
	Params     GameParams  `json:"params"`
	Spec       *GameSpec   `json:"spec,omitempty"`
	Config     *GameConfig `json:"config,omitempty"`
	Status     Status      `json:"status"`
	FailReason string      `json:"fail_reason,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func NewGame(ownerID uuid.UUID, prompt string, params GameParams) (Game, error) {
	if prompt == "" {
		return Game{}, fmt.Errorf("%w: prompt is required", ErrInvalidInput)
	}

	g := Game{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Prompt:    prompt,
		Params:    params,
		Status:    StatusQueued,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return g, nil
}

func (g *Game) StartStage(stage Status) error {
	if !canTransition(g.Status, stage) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, g.Status, stage)
	}

	g.Status = stage
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g *Game) Fail(reason string) error {
	if !canTransition(g.Status, StatusFailed) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, g.Status, StatusFailed)
	}

	g.Status = StatusFailed
	g.FailReason = reason
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g *Game) ApplySpec(spec GameSpec) error {
	g.Spec = &spec
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g *Game) ApplyConfig(cfg GameConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("cfg.Validate: %w", err)
	}

	g.Config = &cfg
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g Game) IsBusy() bool {
	return g.Status != StatusReady && g.Status != StatusFailed
}

func (g Game) IsOwner(userID uuid.UUID) bool {
	return g.OwnerID == userID
}
