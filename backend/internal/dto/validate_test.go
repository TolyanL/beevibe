package dto

import (
	"testing"
)

func TestCreateGameInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateGameInput
		wantErr bool
	}{
		{
			name:  "minimal prompt",
			input: CreateGameInput{Prompt: "змейка"},
		},
		{
			name: "full params",
			input: CreateGameInput{
				Prompt:  "змейка",
				Genre:   "snake",
				Tempo:   tempoPtr(5),
				MapSize: "medium",
				Style:   "neon",
			},
		},
		{
			name:    "empty prompt",
			input:   CreateGameInput{},
			wantErr: true,
		},
		{
			name:    "short prompt",
			input:   CreateGameInput{Prompt: "я"},
			wantErr: true,
		},
		{
			name:    "unknown genre",
			input:   CreateGameInput{Prompt: "змейка", Genre: "arcader"},
			wantErr: true,
		},
		{
			name:    "unknown map size",
			input:   CreateGameInput{Prompt: "змейка", MapSize: "huge"},
			wantErr: true,
		},
		{
			name:    "unknown style",
			input:   CreateGameInput{Prompt: "змейка", Style: "retro"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestRegisterUserInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   RegisterUserInput
		wantErr bool
	}{
		{name: "valid", input: RegisterUserInput{Email: "user@example.com", Password: "password123"}},
		{name: "invalid email", input: RegisterUserInput{Email: "not-an-email", Password: "password123"}, wantErr: true},
		{name: "short password", input: RegisterUserInput{Email: "user@example.com", Password: "short"}, wantErr: true},
		{name: "empty email", input: RegisterUserInput{Password: "password123"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestPatchGameInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   PatchGameInput
		wantErr bool
	}{
		{name: "valid", input: PatchGameInput{Text: "добавь уровень"}},
		{name: "empty text", input: PatchGameInput{}, wantErr: true},
		{name: "single char", input: PatchGameInput{Text: "a"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestRateGameInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "cool", value: "cool"},
		{name: "bad", value: "bad"},
		{name: "unknown", value: "awesome", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := RateGameInput{Value: tt.value}.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestListCatalogInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   ListCatalogInput
		wantErr bool
	}{
		{name: "empty is allowed", input: ListCatalogInput{}},
		{name: "known sort", input: ListCatalogInput{Sort: "cool", Genre: "snake", Limit: 10}},
		{name: "unknown sort", input: ListCatalogInput{Sort: "unknown"}, wantErr: true},
		{name: "unknown genre", input: ListCatalogInput{Genre: "arcader"}, wantErr: true},
		{name: "limit too big", input: ListCatalogInput{Limit: 1000}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestUpdateGameConfigInputValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   UpdateGameConfigInput
		wantErr bool
	}{
		{name: "empty is allowed", input: UpdateGameConfigInput{}},
		{name: "tempo in range", input: UpdateGameConfigInput{Tempo: tempoPtr(7)}},
		{name: "tempo out of range", input: UpdateGameConfigInput{Tempo: tempoPtr(99)}, wantErr: true},
		{name: "unknown map size", input: UpdateGameConfigInput{MapSize: stringPtr("huge")}, wantErr: true},
		{name: "too many colors", input: UpdateGameConfigInput{Palette: []string{"#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777", "#888888"}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func tempoPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
