package domain

import (
	"errors"
	"testing"
)

func TestUserDebit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		balance int64
		amount  int64
		want    int64
		wantErr error
	}{
		{name: "debit from balance", balance: 100, amount: 30, want: 70},
		{name: "debit to zero", balance: 100, amount: 100, want: 0},
		{name: "not enough tokens", balance: 100, amount: 101, wantErr: ErrNotEnoughTokens},
		{name: "zero amount", balance: 100, amount: 0, wantErr: ErrInvalidInput},
		{name: "negative amount", balance: 100, amount: -10, wantErr: ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := User{TokenBalance: tt.balance}

			err := user.Debit(tt.amount)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Debit() error = %v, want %v", err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Debit() error = %v, want nil", err)
			}

			if user.TokenBalance != tt.want {
				t.Fatalf("balance = %d, want %d", user.TokenBalance, tt.want)
			}
		})
	}
}

func TestUserCredit(t *testing.T) {
	t.Parallel()

	user := User{TokenBalance: 10}

	if err := user.Credit(90); err != nil {
		t.Fatalf("Credit() error = %v, want nil", err)
	}

	if user.TokenBalance != 100 {
		t.Fatalf("balance = %d, want %d", user.TokenBalance, 100)
	}

	if err := user.Credit(0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Credit() error = %v, want ErrInvalidInput", err)
	}
}

func TestNewUser(t *testing.T) {
	t.Parallel()

	if _, err := NewUser("not-an-email", "hash", 0); err == nil {
		t.Fatal("NewUser() error = nil, want validation error")
	}

	if _, err := NewUser("user@example.com", "hash", -1); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("NewUser() error = %v, want ErrInvalidInput", err)
	}

	user, err := NewUser("user@example.com", "hash", 100)
	if err != nil {
		t.Fatalf("NewUser() error = %v, want nil", err)
	}

	if user.TokenBalance != 100 {
		t.Fatalf("balance = %d, want %d", user.TokenBalance, 100)
	}
}
