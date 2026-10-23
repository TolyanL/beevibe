package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPasswordVerifiesAgainstOriginal(t *testing.T) {
	t.Parallel()

	a := New(Config{BcryptCost: bcrypt.MinCost})

	hash, err := a.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if hash == "" {
		t.Fatal("hash = empty, want value")
	}

	if err := a.CheckPassword(hash, "password123"); err != nil {
		t.Fatalf("CheckPassword() error = %v, want nil", err)
	}
}

func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	t.Parallel()

	a := New(Config{BcryptCost: bcrypt.MinCost})

	first, err := a.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	second, err := a.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if first == second {
		t.Error("hashes are equal, want different salts")
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	a := New(Config{BcryptCost: bcrypt.MinCost})

	hash, err := a.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if err := a.CheckPassword(hash, "wrong-password"); err == nil {
		t.Fatal("CheckPassword() error = nil, want error")
	}
}

func TestCheckPasswordRejectsMalformedHash(t *testing.T) {
	t.Parallel()

	a := New(Config{BcryptCost: bcrypt.MinCost})

	if err := a.CheckPassword("not-a-hash", "password123"); err == nil {
		t.Fatal("CheckPassword() error = nil, want error")
	}
}

func TestGenerateTokenReturnsUniqueValues(t *testing.T) {
	t.Parallel()

	a := New(Config{BcryptCost: bcrypt.MinCost})

	first, err := a.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v, want nil", err)
	}

	second, err := a.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v, want nil", err)
	}

	if first == "" || second == "" {
		t.Fatal("token = empty, want value")
	}

	if first == second {
		t.Error("tokens are equal, want different values")
	}

	if len(first) != 64 {
		t.Errorf("token length = %d, want 64", len(first))
	}
}
