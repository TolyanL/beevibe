package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string      `envconfig:"KAFKA_BROKERS"       default:"localhost:9094"`
	Topic        string        `envconfig:"KAFKA_TOPIC"         default:"game.events"`
	GroupID      string        `envconfig:"KAFKA_GROUP_ID"      default:"gamegen"`
	WriteTimeout time.Duration `envconfig:"KAFKA_WRITE_TIMEOUT" default:"10s"`
	ReadTimeout  time.Duration `envconfig:"KAFKA_READ_TIMEOUT"  default:"10s"`
}

// Producer — реализация порта usecase/outbox.Kafka.
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
