package validator

import (
	"context"
	"fmt"
	"math"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (v *Validator) ExtractConfig(ctx context.Context, html string) (domain.GameConfig, error) {
	ctx, span := tracer.Start(ctx, "adapter validator ExtractConfig")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return domain.GameConfig{}, fmt.Errorf("validator.ExtractConfig: %w", err)
	}

	cfg, err := extractConfig(html)
	if err != nil {
		return domain.GameConfig{}, fmt.Errorf("validator.ExtractConfig: %w", err)
	}

	return cfg, nil
}

func extractConfig(html string) (domain.GameConfig, error) {
	start, end, ok := findConfigObject(html)
	if !ok {
		return domain.GameConfig{}, fmt.Errorf("%w: config block not found", domain.ErrValidationFailed)
	}

	obj, err := parseJSObject(html[start : end+1])
	if err != nil {
		return domain.GameConfig{}, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	cfg := domain.GameConfig{}

	tempo, err := requiredNumberMember(obj, "tempo")
	if err != nil {
		return domain.GameConfig{}, err
	}

	if tempo != math.Trunc(tempo) {
		return domain.GameConfig{}, fmt.Errorf("%w: tempo must be an integer", domain.ErrValidationFailed)
	}

	cfg.Tempo = domain.Tempo(int(tempo))

	palette, err := optionalStringsMember(obj, "palette")
	if err != nil {
		return domain.GameConfig{}, err
	}

	cfg.Palette = palette

	mapSize, err := requiredStringMember(obj, "mapSize", "map_size")
	if err != nil {
		return domain.GameConfig{}, err
	}

	cfg.MapSize = domain.MapSize(mapSize)

	difficulty, err := optionalStringMember(obj, "difficulty")
	if err != nil {
		return domain.GameConfig{}, err
	}

	cfg.Difficulty = difficulty

	controls, err := optionalStringMember(obj, "controls", "control")
	if err != nil {
		return domain.GameConfig{}, err
	}

	cfg.Controls = controls

	style, err := requiredStringMember(obj, "style")
	if err != nil {
		return domain.GameConfig{}, err
	}

	cfg.Style = domain.Style(style)

	if err := cfg.Validate(); err != nil {
		return domain.GameConfig{}, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	return cfg, nil
}

func requiredNumberMember(obj *jsValue, names ...string) (float64, error) {
	member := obj.member(names...)
	if member == nil {
		return 0, fmt.Errorf("%w: %s is missing", domain.ErrValidationFailed, names[0])
	}

	if member.value.kind != jsNumber {
		return 0, fmt.Errorf("%w: %s must be a number", domain.ErrValidationFailed, names[0])
	}

	return member.value.num, nil
}

func requiredStringMember(obj *jsValue, names ...string) (string, error) {
	member := obj.member(names...)
	if member == nil {
		return "", fmt.Errorf("%w: %s is missing", domain.ErrValidationFailed, names[0])
	}

	if member.value.kind != jsString {
		return "", fmt.Errorf("%w: %s must be a string", domain.ErrValidationFailed, names[0])
	}

	return member.value.str, nil
}

func optionalStringMember(obj *jsValue, names ...string) (string, error) {
	member := obj.member(names...)
	if member == nil {
		return "", nil
	}

	if member.value.kind != jsString {
		return "", fmt.Errorf("%w: %s must be a string", domain.ErrValidationFailed, names[0])
	}

	return member.value.str, nil
}

func optionalStringsMember(obj *jsValue, names ...string) ([]string, error) {
	member := obj.member(names...)
	if member == nil {
		return nil, nil
	}

	if member.value.kind != jsArray {
		return nil, fmt.Errorf("%w: %s must be an array of strings", domain.ErrValidationFailed, names[0])
	}

	values := make([]string, 0, len(member.value.elems))

	for _, elem := range member.value.elems {
		if elem.kind != jsString {
			return nil, fmt.Errorf("%w: %s must be an array of strings", domain.ErrValidationFailed, names[0])
		}

		values = append(values, elem.str)
	}

	return values, nil
}
