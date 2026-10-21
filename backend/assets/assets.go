package assets

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"strings"
)

//go:embed prompts/*.md
var fs embed.FS

const (
	promptsDir    = "prompts"
	promptExt     = ".md"
	versionLength = 16
)

// Prompt — текст системного промпта по имени без расширения.
func Prompt(name string) (string, error) {
	data, err := fs.ReadFile(promptsDir + "/" + name + promptExt)
	if err != nil {
		return "", fmt.Errorf("assets.Prompt: %w", err)
	}

	return string(data), nil
}

// PromptVersion — короткий хэш содержимого всех промптов: меняется при правке любого из них.
func PromptVersion() string {
	hash := sha256.New()

	for _, name := range PromptNames() {
		data, err := fs.ReadFile(promptsDir + "/" + name + promptExt)
		if err != nil {
			continue
		}

		hash.Write([]byte(name))
		hash.Write(data)
	}

	return hex.EncodeToString(hash.Sum(nil))[:versionLength]
}

// PromptNames — имена доступных промптов без расширения, по алфавиту.
func PromptNames() []string {
	entries, err := fs.ReadDir(promptsDir)
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
