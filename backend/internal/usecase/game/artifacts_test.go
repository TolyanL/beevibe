package game

import (
	"testing"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestResumeStageStartsFromTheStageAfterTheLastDone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []domain.GenEvent
		want   domain.Status
	}{
		{name: "no events", want: domain.StatusAnalyzing},
		{
			name: "failed at analyze",
			events: []domain.GenEvent{
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"error"}`)},
			},
			want: domain.StatusAnalyzing,
		},
		{
			name: "failed at coding",
			events: []domain.GenEvent{
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"planning","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"coding","status":"error"}`)},
			},
			want: domain.StatusCoding,
		},
		{
			name: "failed at assembling",
			events: []domain.GenEvent{
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"planning","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"coding","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"assets","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"reviewing","status":"done"}`)},
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"assembling","status":"error"}`)},
			},
			want: domain.StatusAssembling,
		},
		{
			name: "all stages done restarts from the beginning",
			events: []domain.GenEvent{
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"checking","status":"done"}`)},
			},
			want: domain.StatusAnalyzing,
		},
		{
			name: "non-pipeline stages are ignored",
			events: []domain.GenEvent{
				{Type: domain.GenEventStage, Payload: []byte(`{"stage":"meta","status":"done","operation":"meta"}`)},
			},
			want: domain.StatusAnalyzing,
		},
	}

	for _, tt := range tests {
		if got := resumeStage(tt.events); got != tt.want {
			t.Errorf("%s: resumeStage() = %s, want %s", tt.name, got, tt.want)
		}
	}
}
