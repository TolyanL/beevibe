package llm

import (
	"fmt"
)

func (c *Client) Prompt(name string) (string, error) {
	text, err := Prompt(name)
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	return text, nil
}

func (c *Client) PromptVersion() string {
	return PromptVersion()
}
