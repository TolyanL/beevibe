package llm

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"strings"
)

//go:embed prompts/*.md
var promptsFS embed.FS

const (
	promptsDir    = "prompts"
	promptExt     = ".md"
	promptVersion = 16
)

func Prompt(name string) (string, error) {
	data, err := promptsFS.ReadFile(promptsDir + "/" + name + promptExt)
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	return string(data), nil
}

func PromptVersion() string {
	hash := sha256.New()

	for _, name := range PromptNames() {
		data, err := promptsFS.ReadFile(promptsDir + "/" + name + promptExt)
		if err != nil {
			continue
		}

		hash.Write([]byte(name))
		hash.Write(data)
	}

	return hex.EncodeToString(hash.Sum(nil))[:promptVersion]
}

func PromptNames() []string {
	entries, err := promptsFS.ReadDir(promptsDir)
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), promptExt)
		if name == entry.Name() {
			continue
		}

		names = append(names, name)
	}

	return names
}
