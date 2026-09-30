package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Transcriber interface {
	Transcribe(ctx context.Context, request TranscriptionRequest) error
}

type TranscriptionRequest struct {
	AudioPath  string
	SourcePath string
	Model      string
	Language   string
}

func ensureSourceDir(sourcePath string) error {
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		return fmt.Errorf("create transcript directory: %w", err)
	}
	return nil
}
