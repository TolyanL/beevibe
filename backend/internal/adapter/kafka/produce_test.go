package kafka

import (
	"context"
	"errors"
	"testing"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/kafka/mocks"
)

func newTestProducer(t *testing.T) (*Producer, *mocks.MockMessageWriter) {
	t.Helper()

	writer := mocks.NewMockMessageWriter(t)

	return &Producer{writer: writer}, writer
}

func TestProduceSendsMessageWithTopicKeyAndPayload(t *testing.T) {
	t.Parallel()

	producer, writer := newTestProducer(t)

	payload := []byte(`{"id":"event-1"}`)

	writer.EXPECT().WriteMessages(mock.Anything, mock.MatchedBy(func(msgs []kafkago.Message) bool {
		return len(msgs) == 1 &&
			msgs[0].Topic == "game.events" &&
			string(msgs[0].Key) == "game-1" &&
			string(msgs[0].Value) == string(payload)
	})).Return(nil)

	if err := producer.Produce(context.Background(), "game.events", "game-1", payload); err != nil {
		t.Fatalf("Produce() error = %v, want nil", err)
	}
}

func TestProduceWrapsWriterError(t *testing.T) {
	t.Parallel()

	producer, writer := newTestProducer(t)

	cause := errors.New("broker is down")

	writer.EXPECT().WriteMessages(mock.Anything, mock.Anything).Return(cause)

	err := producer.Produce(context.Background(), "game.events", "game-1", nil)
	if !errors.Is(err, cause) {
		t.Fatalf("Produce() error = %v, want wrapped writer error", err)
	}

	if got := err.Error(); got != "kafka.Writer.WriteMessages: broker is down" {
		t.Errorf("error = %q, want wrapped kafka message", got)
	}
}

func TestCloseClosesWriter(t *testing.T) {
	t.Parallel()

	producer, writer := newTestProducer(t)

	writer.EXPECT().Close().Return(nil)

	if err := producer.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

func TestCloseWrapsWriterError(t *testing.T) {
	t.Parallel()

	producer, writer := newTestProducer(t)

	writer.EXPECT().Close().Return(errors.New("close failed"))

	if err := producer.Close(); err == nil {
		t.Fatal("Close() error = nil, want error")
	}
}

func TestNewBuildsWriterFromConfig(t *testing.T) {
	t.Parallel()

	producer := New(Config{
		Brokers:      []string{"localhost:9094"},
		Topic:        "game.events",
		GroupID:      "beevibe",
		WriteTimeout: 0,
		ReadTimeout:  0,
	})
	if producer == nil || producer.writer == nil {
		t.Fatal("New() = nil, want producer with writer")
	}
}
