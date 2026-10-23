package tracer

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type Span interface {
	End()
	SetError(err error)
	SetAttributes(attrs map[string]string)
}

type spanKey struct{}

type span struct {
	name  string
	start time.Time
	attrs map[string]string
	err   error
}

func Start(ctx context.Context, name string) (context.Context, Span) {
	s := &span{name: name, start: time.Now(), attrs: map[string]string{}}

	return context.WithValue(ctx, spanKey{}, s), s
}

func (s *span) End() {
	event := log.Debug().Str("span", s.name).Dur("duration", time.Since(s.start))

	for key, value := range s.attrs {
		event = event.Str(key, value)
	}

	if s.err != nil {
		event = event.Err(s.err)
	}

	event.Msg("span")
}

func (s *span) SetError(err error) {
	s.err = err
}

func (s *span) SetAttributes(attrs map[string]string) {
	for key, value := range attrs {
		s.attrs[key] = value
	}
}
