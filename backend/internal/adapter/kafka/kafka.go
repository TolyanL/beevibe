package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string      `yaml:"brokers"`
	Topic        string        `yaml:"topic"`
	GroupID      string        `yaml:"group_id"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
}

// NOTE: the writer is an interface so the producer can be tested without a broker.
type MessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer struct {
	writer MessageWriter
}

func New(cfg Config) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Balancer:               &kafka.LeastBytes{},
			WriteTimeout:           cfg.WriteTimeout,
			AllowAutoTopicCreation: true,
		},
	}
}
