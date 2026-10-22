package llm

import (
	"fmt"
)

// Prompt — текст системного промпта из встроенных ассетов.
func (c *Client) Prompt(name string) (string, error) {
	text, err := Prompt(name)
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	return text, nil
}

// PromptVersion — короткий хэш версии всех встроенных промптов.
func (c *Client) PromptVersion() string {
	return PromptVersion()
}
