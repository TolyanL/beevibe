package validator

import (
	"os/exec"
)

type Config struct {
	// NOTE: пустой NodePath означает поиск node в PATH.
	NodePath string `envconfig:"NODE_PATH" default:""`
}

type Validator struct {
	nodePath string
}

func New(cfg Config) *Validator {
	// NOTE: отсутствие node не является ошибкой: проверка синтаксиса JS
	// в этом случае пропускается.
	nodePath := cfg.NodePath

	if nodePath == "" {
		if path, err := exec.LookPath("node"); err == nil {
			nodePath = path
		}
	}

	return &Validator{nodePath: nodePath}
}
