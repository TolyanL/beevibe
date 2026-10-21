package llm

import (
	"fmt"

	"github.com/TolyanL/beevibe/assets"
)

// Prompt — текст системного промпта из встроенных ассетов.
func (c *Client) Prompt(name string) (string, error) {
	text, err := assets.Prompt(name)
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	return text, nil
}

// PromptVersion — короткий хэш версии всех встроенных промптов.
func (c *Client) PromptVersion() string {
	return assets.PromptVersion()
}
