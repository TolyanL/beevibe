package app

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/TolyanL/beevibe/internal/usecase"
)

const workerPollInterval = 300 * time.Millisecond

type Worker struct {
	usecase     *usecase.UseCase
	concurrency int
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewWorker(uc *usecase.UseCase, concurrency int) *Worker {
	return &Worker{
		usecase:     uc,
		concurrency: concurrency,
		stop:        make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	for range w.concurrency {
		w.wg.Go(func() {
			w.loop(ctx)
		})
	}
}

func (w *Worker) loop(ctx context.Context) {
	ticker := time.NewTicker(workerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			if _, err := w.usecase.RunNextQueuedGame(ctx); err != nil {
				log.Error().Err(err).Msg("generation failed")
			}
		}
	}
}

func (w *Worker) Close() {
	close(w.stop)
	w.wg.Wait()
}
