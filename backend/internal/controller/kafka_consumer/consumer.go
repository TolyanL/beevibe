package kafka_consumer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/TolyanL/beevibe/internal/adapter/kafka"
)

type Handler interface {
	HandleEvent(ctx context.Context, topic, key string, payload []byte) error
}

// NOTE: the reader is an interface so the fetch/commit loop can be tested
// without a broker.
type MessageReader interface {
	FetchMessage(ctx context.Context) (kafkago.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

type Consumer struct {
	reader  MessageReader
	handler Handler
	stop    chan struct{}
	wg      sync.WaitGroup
}

func New(cfg kafka.Config, handler Handler) *Consumer {
	return &Consumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:     cfg.Brokers,
			Topic:       cfg.Topic,
			GroupID:     cfg.GroupID,
			MaxWait:     cfg.ReadTimeout,
			StartOffset: kafkago.LastOffset,
		}),
		handler: handler,
		stop:    make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.wg.Go(func() {
		c.loop(ctx)
	})
}

func (c *Consumer) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stop:
			return
		default:
		}

		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			log.Error().Err(err).Msg("kafka: fetch message failed")

			select {
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			case <-time.After(time.Second):
			}

			continue
		}

		if err := c.handler.HandleEvent(ctx, message.Topic, string(message.Key), message.Value); err != nil {
			log.Error().Err(err).Msg("kafka: handle event failed")

			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			log.Error().Err(err).Msg("kafka: commit message failed")
		}
	}
}

func (c *Consumer) Close() error {
	close(c.stop)
	c.wg.Wait()

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("kafka.Reader.Close: %w", err)
	}

	return nil
}
