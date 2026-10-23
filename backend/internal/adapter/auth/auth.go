package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	BcryptCost int `yaml:"bcrypt_cost"`
}

type Auth struct {
	cost int
}

func New(cfg Config) *Auth {
	return &Auth{cost: cfg.BcryptCost}
}

func (a *Auth) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), a.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	return string(hash), nil
}

func (a *Auth) CheckPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err)
	}

	return nil
}

func (a *Auth) GenerateToken() (string, error) {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return hex.EncodeToString(buf), nil
}
