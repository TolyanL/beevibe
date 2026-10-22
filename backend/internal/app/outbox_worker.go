package app

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/TolyanL/beevibe/internal/usecase"
)

const (
	outboxPollInterval = time.Second
	outboxBatchSize    = 50
)

// OutboxWorker — фоновая публикация накопленных событий в Kafka.
type OutboxWorker struct {
	usecase  *usecase.UseCase
	interval time.Duration
	stop     chan struct{}
	wg       sync.WaitGroup
}

func NewOutboxWorker(uc *usecase.UseCase, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{
		usecase:  uc,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	w.wg.Go(func() {
		w.loop(ctx)
	})
}

func (w *OutboxWorker) loop(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			if _, err := w.usecase.Outbox.PublishBatch(ctx, outboxBatchSize); err != nil {
				log.Error().Err(err).Msg("outbox publish failed")
			}
		}
	}
}

func (w *OutboxWorker) Close() {
	close(w.stop)
	w.wg.Wait()
}
