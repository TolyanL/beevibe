package static

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const indexHTML = "<html>spa</html>"

func newSPA(t *testing.T) http.Handler {
	t.Helper()

	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}

	files := map[string]string{
		"index.html":    indexHTML,
		"assets/app.js": "console.log(1)",
	}

	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}
	}

	handler, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return handler
}

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	handler := newSPA(t)

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{
			name:       "index",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
			wantCache:  "no-cache",
		},
		{
			name:       "spa fallback",
			method:     http.MethodGet,
			target:     "/games/42",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
			wantCache:  "no-cache",
		},
		{
			name:       "asset",
			method:     http.MethodGet,
			target:     "/assets/app.js",
			wantStatus: http.StatusOK,
			wantBody:   "console.log(1)",
			wantCache:  "public, max-age=31536000, immutable",
		},
		{
			name:       "traversal is rejected",
			method:     http.MethodGet,
			target:     "/../../etc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "post is not found",
			method:     http.MethodPost,
			target:     "/catalog",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.target, nil)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				if body := recorder.Body.String(); body != tt.wantBody {
					t.Fatalf("body = %q, want %q", body, tt.wantBody)
				}
			}

			if cache := recorder.Header().Get("Cache-Control"); cache != tt.wantCache {
				t.Fatalf("cache-control = %q, want %q", cache, tt.wantCache)
			}
		})
	}
}

func TestNewMissingDir(t *testing.T) {
	t.Parallel()

	if _, err := New(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("New: expected an error for a missing directory")
	}
}
