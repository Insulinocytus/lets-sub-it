package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Transcriber interface {
	Transcribe(ctx context.Context, mediaPath string) ([]Segment, error)
}

type fakeTranscriber struct {
	segments []Segment
}

func FakeTranscriber(segments []Segment) Transcriber {
	cloned := make([]Segment, len(segments))
	copy(cloned, segments)
	return fakeTranscriber{segments: cloned}
}

func (f fakeTranscriber) Transcribe(_ context.Context, _ string) ([]Segment, error) {
	cloned := make([]Segment, len(f.segments))
	copy(cloned, f.segments)
	return cloned, nil
}

type commandTranscriber struct {
	command string
}

func NewFastWhisperTranscriber() Transcriber {
	return commandTranscriber{command: strings.TrimSpace(os.Getenv("FAST_WHISPER_COMMAND"))}
}

func (t commandTranscriber) Transcribe(ctx context.Context, mediaPath string) ([]Segment, error) {
	if t.command == "" {
		return nil, errors.New("FAST_WHISPER_COMMAND is required for worker transcription")
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", t.command+" \"$1\"", "fast-whisper", mediaPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("fast-whisper command failed: %w: %s", err, string(output))
	}

	var payload []struct {
		Start string `json:"start"`
		End   string `json:"end"`
		Text  string `json:"text"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("decode fast-whisper output: %w", err)
	}

	segments := make([]Segment, len(payload))
	for i, item := range payload {
		segments[i] = Segment{
			Start:      item.Start,
			End:        item.End,
			SourceText: item.Text,
		}
	}

	return segments, nil
}
