package domain

import (
	"fmt"
	"strings"
	"time"

	"unicode"

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
	StyleCustom  Style = "custom"
)

const (
	CustomLabelMaxLength = 32
	customLabelMinLength = 2
)

type CustomLabel string

type GameParams struct {
	Genre       Genre       `json:"genre,omitempty"`
	Tempo       Tempo       `json:"tempo,omitempty"`
	MapSize     MapSize     `json:"map_size,omitempty"`
	Style       Style       `json:"style,omitempty"`
	CustomGenre CustomLabel `json:"custom_genre,omitempty"`
	CustomStyle CustomLabel `json:"custom_style,omitempty"`
}

func (p GameParams) Validate() error {
	if p.Genre == GenreCustom {
		if err := p.CustomGenre.validate(); err != nil {
			return fmt.Errorf("custom genre: %w", err)
		}
	}

	if p.Style == StyleCustom {
		if err := p.CustomStyle.validate(); err != nil {
			return fmt.Errorf("custom style: %w", err)
		}
	}

	return nil
}

func (l CustomLabel) validate() error {
	value := strings.TrimSpace(string(l))

	length := len([]rune(value))
	if length < customLabelMinLength || length > CustomLabelMaxLength {
		return fmt.Errorf("%w: label must be %d..%d characters", ErrInvalidInput, customLabelMinLength, CustomLabelMaxLength)
	}

	for _, r := range value {
		isLetter := unicode.IsLetter(r)
		isDigit := unicode.IsDigit(r)

		if !isLetter && !isDigit && r != ' ' && r != '-' {
			return fmt.Errorf("%w: label may contain letters, digits, spaces and hyphens only", ErrInvalidInput)
		}
	}

	return nil
}

func NewCustomLabel(value string) (CustomLabel, error) {
	label := CustomLabel(strings.TrimSpace(value))

	if strings.TrimSpace(value) == "" {
		return "", nil
	}

	if err := label.validate(); err != nil {
		return "", fmt.Errorf("label.validate: %w", err)
	}

	return label, nil
}

type Tempo int

const (
	TempoSlow Tempo = 1
	TempoFast Tempo = 10
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type Game struct {
	ID          uuid.UUID   `json:"id"`
	OwnerID     uuid.UUID   `json:"owner_id"`
	Prompt      string      `json:"prompt"`
	Params      GameParams  `json:"params"`
	Spec        *GameSpec   `json:"spec,omitempty"`
	Config      *GameConfig `json:"config,omitempty"`
	Status      Status      `json:"status"`
	FailReason  string      `json:"fail_reason,omitempty"`
	Visibility  Visibility  `json:"visibility"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty"`
	Title       string      `json:"title"`
	Slug        string      `json:"slug"`
	Description string      `json:"description"`
	Cover       *GameCover  `json:"cover,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func NewGame(ownerID uuid.UUID, prompt string, params GameParams) (Game, error) {
	if prompt == "" {
		return Game{}, fmt.Errorf("%w: prompt is required", ErrInvalidInput)
	}

	id := uuid.New()

	if err := params.Validate(); err != nil {
		return Game{}, fmt.Errorf("params.Validate: %w", err)
	}

	g := Game{
		ID:         id,
		Slug:       NewGameSlug(prompt, id),
		OwnerID:    ownerID,
		Prompt:     prompt,
		Params:     params,
		Status:     StatusQueued,
		Visibility: VisibilityPrivate,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
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

func (g *Game) SetMeta(title, description string, cover *GameCover) error {
	meta, err := NewGameMeta(title, description, cover)
	if err != nil {
		return fmt.Errorf("NewGameMeta: %w", err)
	}

	g.Title = meta.Title
	g.Description = meta.Description
	g.Cover = meta.Cover
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g Game) DisplaySlug() string {
	if g.Slug != "" {
		return g.Slug
	}

	return NewGameSlug(g.DisplayTitle(), g.ID)
}

func (g Game) DisplayTitle() string {
	if title := strings.TrimSpace(g.Title); title != "" {
		return title
	}

	return g.Prompt
}

func (g *Game) Retry() error {
	if g.Status != StatusFailed {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, g.Status, StatusQueued)
	}

	g.Status = StatusQueued
	g.FailReason = ""
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

func (g *Game) Publish() error {
	if g.Status != StatusReady {
		return fmt.Errorf("%w: only a ready game can be published", ErrInvalidTransition)
	}

	g.Visibility = VisibilityPublic
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g *Game) Unpublish() error {
	if g.Visibility == VisibilityPrivate {
		return fmt.Errorf("%w: game is already private", ErrInvalidTransition)
	}

	g.Visibility = VisibilityPrivate
	g.UpdatedAt = time.Now().UTC()

	return nil
}

func (g *Game) Delete() error {
	if g.IsDeleted() {
		return fmt.Errorf("%w: game is already deleted", ErrInvalidTransition)
	}

	if g.Status == StatusPatching {
		return fmt.Errorf("%w: an edit is in progress", ErrGenerationInProgress)
	}

	now := time.Now().UTC()
	g.DeletedAt = &now
	g.UpdatedAt = now

	return nil
}

func (g Game) IsDeleted() bool {
	return g.DeletedAt != nil
}

func (g Game) IsPublic() bool {
	return g.Visibility == VisibilityPublic && g.Status == StatusReady && !g.IsDeleted()
}

func (g Game) IsBusy() bool {
	return g.Status != StatusReady && g.Status != StatusFailed
}

func (g Game) IsOwner(userID uuid.UUID) bool {
	return g.OwnerID == userID
}
