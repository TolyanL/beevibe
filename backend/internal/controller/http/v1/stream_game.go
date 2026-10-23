package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

const (
	streamPollInterval = 300 * time.Millisecond
	streamHeartbeat    = 15 * time.Second
)

func (h *Handlers) StreamGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, errUnauthorized)

		return
	}

	gameID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, errInvalidBody)

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, errors.New("streaming is not supported"))

		return
	}

	afterSeq := parseLastEventID(r.Header.Get("Last-Event-ID"))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	poll := time.NewTicker(streamPollInterval)
	defer poll.Stop()

	heartbeat := time.NewTicker(streamHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case <-poll.C:
			output, err := h.usecase.Game.ListGenEvents(r.Context(), dto.ListGenEventsInput{
				UserID:   userID,
				GameID:   gameID,
				AfterSeq: afterSeq,
			})
			if err != nil {
				writeStreamError(w, flusher, err)

				return
			}

			for _, event := range output.Events {
				fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.Type, event.Payload)
				afterSeq = event.Seq
			}

			flusher.Flush()

			if len(output.Events) == 0 && (output.Status == string(domain.StatusReady) || output.Status == string(domain.StatusFailed)) {
				return
			}
		}
	}
}

func writeStreamError(w http.ResponseWriter, flusher http.Flusher, err error) {
	fmt.Fprintf(w, "event: error\ndata: {\"code\":\"stream_failed\",\"message\":%q}\n\n", err.Error())
	flusher.Flush()
}

func parseLastEventID(value string) int64 {
	if value == "" {
		return 0
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return parsed
}
