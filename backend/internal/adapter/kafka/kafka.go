package kafka

import (
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

type Producer struct {
	writer *kafka.Writer
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
