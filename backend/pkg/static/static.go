package static

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const indexFile = "index.html"

type spa struct {
	dir   string
	files http.Handler
}

func New(dir string) (http.Handler, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("os.Stat: %s is not a directory", dir)
	}

	if _, err := os.Stat(filepath.Join(dir, indexFile)); err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}

	return &spa{dir: dir, files: http.FileServer(http.Dir(dir))}, nil
}

func (s *spa) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)

		return
	}

	clean := path.Clean("/" + r.URL.Path)
	file := filepath.Join(s.dir, filepath.FromSlash(clean))

	if info, err := os.Stat(file); err == nil && !info.IsDir() {
		if strings.HasPrefix(clean, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		s.files.ServeHTTP(w, r)

		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, filepath.Join(s.dir, indexFile))
}
