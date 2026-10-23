package validator

import (
	"os/exec"
)

type Config struct {
	NodePath string `envconfig:"NODE_PATH" default:""`
}

type Validator struct {
	nodePath string
}

func New(cfg Config) *Validator {
	nodePath := cfg.NodePath

	if nodePath == "" {
		if path, err := exec.LookPath("node"); err == nil {
			nodePath = path
		}
	}

	return &Validator{nodePath: nodePath}
}
