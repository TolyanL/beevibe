package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Port            string        `envconfig:"HTTP_PORT"             default:"8080"`
	ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT"     default:"30s"`
	WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"    default:"0s"`
	IdleTimeout     time.Duration `envconfig:"HTTP_IDLE_TIMEOUT"     default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"10s"`
}

type Server struct {
	srv *http.Server
}

func New(cfg Config, handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("srv.ListenAndServe: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("srv.Shutdown: %w", err)
	}

	return nil
}
