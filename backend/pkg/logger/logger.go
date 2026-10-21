package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level  string `envconfig:"LOG_LEVEL"  default:"info"`
	Format string `envconfig:"LOG_FORMAT" default:"console"`
}

type ContextErrKey struct{}

func Init(cfg Config) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	if cfg.Format == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		return
	}

	writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctxErr := new(error)
		ctx = context.WithValue(ctx, ContextErrKey{}, ctxErr)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(rec, r.WithContext(ctx))
		duration := time.Since(start)

		event := log.Info()
		if *ctxErr != nil {
			event = log.Error().Err(*ctxErr)
		}

		if rec.status >= http.StatusInternalServerError {
			event = log.Error().Int("status", rec.status)
		}

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.status).
			Dur("duration", duration).
			Msg("http request")
	})
}

type statusRecorder struct {
	http.ResponseWriter

	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Flush — проксирование flush'а для SSE: без него http.Fluster не доступен.
func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
