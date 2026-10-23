package validator

import (
	"os/exec"
)

type Config struct {
	NodePath string `yaml:"node_path"`
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
