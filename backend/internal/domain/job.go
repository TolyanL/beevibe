package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type JobKind string

const (
	JobKindPatch JobKind = "patch"
	JobKindFix   JobKind = "fix"
	JobKindMeta  JobKind = "meta"
)

const (
	jobInstructionMaxLength = 4000
	jobRuntimeErrorMaxLen   = 2000
)

type GameJob struct {
	ID          uuid.UUID `json:"id"`
	GameID      uuid.UUID `json:"game_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Kind        JobKind   `json:"kind"`
	Instruction string    `json:"instruction"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewGameJob(gameID, ownerID uuid.UUID, kind JobKind, instruction string) (GameJob, error) {
	if !kind.IsValid() {
		return GameJob{}, fmt.Errorf("%w: unknown job kind %q", ErrInvalidInput, kind)
	}

	instruction = strings.TrimSpace(instruction)

	if len([]rune(instruction)) > jobInstructionMaxLength {
		return GameJob{}, fmt.Errorf("%w: instruction is too long", ErrInvalidInput)
	}

	if kind == JobKindPatch && instruction == "" {
		return GameJob{}, fmt.Errorf("%w: patch instruction is required", ErrInvalidInput)
	}

	if len([]rune(instruction)) > jobRuntimeErrorMaxLen && kind == JobKindFix {
		instruction = string([]rune(instruction)[:jobRuntimeErrorMaxLen])
	}

	return GameJob{
		ID:          uuid.New(),
		GameID:      gameID,
		OwnerID:     ownerID,
		Kind:        kind,
		Instruction: instruction,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (k JobKind) IsValid() bool {
	return k == JobKindPatch || k == JobKindFix || k == JobKindMeta
}

func (k JobKind) String() string {
	return string(k)
}
