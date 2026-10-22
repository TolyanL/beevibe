package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func (p *Producer) Produce(ctx context.Context, topic, key string, payload []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("kafka.Writer.WriteMessages: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("kafka.Writer.Close: %w", err)
	}

	return nil
}
