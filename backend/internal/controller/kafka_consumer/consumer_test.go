package kafka_consumer

import (
	"context"
	"errors"
	"testing"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/controller/kafka_consumer/mocks"
)

func newTestConsumer(t *testing.T) (*Consumer, *mocks.MockMessageReader, *mocks.MockHandler) {
	t.Helper()

	reader := mocks.NewMockMessageReader(t)
	handler := mocks.NewMockHandler(t)

	consumer := &Consumer{
		reader:  reader,
		handler: handler,
		stop:    make(chan struct{}),
	}

	return consumer, reader, handler
}

func TestLoopHandlesMessageAndCommitsOffset(t *testing.T) {
	t.Parallel()

	consumer, reader, handler := newTestConsumer(t)

	message := kafkago.Message{
		Topic: "game.events",
		Key:   []byte("game-1"),
		Value: []byte(`{"id":"event-1"}`),
	}

	reader.EXPECT().FetchMessage(mock.Anything).Return(message, nil).Once()
	handler.EXPECT().HandleEvent(mock.Anything, "game.events", "game-1", message.Value).Return(nil)
	reader.EXPECT().CommitMessages(mock.Anything, []kafkago.Message{message}).Return(nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, context.Canceled)

	consumer.loop(context.Background())
}

func TestLoopSkipsCommitWhenHandlerFails(t *testing.T) {
	t.Parallel()

	consumer, reader, handler := newTestConsumer(t)

	message := kafkago.Message{Topic: "game.events", Key: []byte("game-1"), Value: []byte(`{}`)}

	reader.EXPECT().FetchMessage(mock.Anything).Return(message, nil).Once()
	handler.EXPECT().HandleEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("handler failed"))
	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, context.Canceled)

	consumer.loop(context.Background())
}

func TestLoopRetriesAfterFetchError(t *testing.T) {
	t.Parallel()

	consumer, reader, handler := newTestConsumer(t)

	message := kafkago.Message{Topic: "game.events", Key: []byte("game-1"), Value: []byte(`{}`)}

	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, errors.New("temporary")).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(message, nil).Once()
	handler.EXPECT().HandleEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	reader.EXPECT().CommitMessages(mock.Anything, []kafkago.Message{message}).Return(nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, context.Canceled)

	consumer.loop(context.Background())
}

func TestLoopKeepsRunningWhenCommitFails(t *testing.T) {
	t.Parallel()

	consumer, reader, handler := newTestConsumer(t)

	first := kafkago.Message{Topic: "game.events", Key: []byte("game-1"), Value: []byte(`{}`)}
	second := kafkago.Message{Topic: "game.events", Key: []byte("game-2"), Value: []byte(`{}`)}

	reader.EXPECT().FetchMessage(mock.Anything).Return(first, nil).Once()
	handler.EXPECT().HandleEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	reader.EXPECT().CommitMessages(mock.Anything, []kafkago.Message{first}).Return(errors.New("commit failed")).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(second, nil).Once()
	handler.EXPECT().HandleEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	reader.EXPECT().CommitMessages(mock.Anything, []kafkago.Message{second}).Return(nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, context.Canceled)

	consumer.loop(context.Background())
}

func TestLoopStopsOnCanceledFetch(t *testing.T) {
	t.Parallel()

	consumer, reader, _ := newTestConsumer(t)

	reader.EXPECT().FetchMessage(mock.Anything).Return(kafkago.Message{}, context.DeadlineExceeded)

	consumer.loop(context.Background())
}

func TestLoopStopsOnClosedStopChannel(t *testing.T) {
	t.Parallel()

	consumer, _, _ := newTestConsumer(t)

	close(consumer.stop)

	consumer.loop(context.Background())
}

func TestLoopStopsWhenContextIsDone(t *testing.T) {
	t.Parallel()

	consumer, _, _ := newTestConsumer(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	consumer.loop(ctx)
}
