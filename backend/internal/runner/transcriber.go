package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Transcriber interface {
	Transcribe(ctx context.Context, request TranscriptionRequest) error
}

type TranscriptionRequest struct {
	JobID       string
	AudioPath   string
	SourcePath  string
	Model       string
	ComputeType string
	Language    string
	OnProgress  func(text string) error
}

func ensureSourceDir(sourcePath string) error {
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		return fmt.Errorf("create transcript directory: %w", err)
	}
	return nil
}

func ensureValidSourceVTT(sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source.vtt was not created: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("empty source.vtt")
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("read source.vtt: %w", err)
	}
	defer file.Close()

	const prefix = "WEBVTT"
	data := make([]byte, len(prefix))
	n, err := io.ReadFull(file, data)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return fmt.Errorf("read source.vtt: %w", err)
	}
	if string(data[:n]) != prefix {
		return fmt.Errorf("source.vtt must start with WEBVTT")
	}
	return nil
}
