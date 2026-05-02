package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func transcribeAudio(ctx context.Context, audioPath string, sourcePath string, model string, computeType string, language string) error {
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		return fmt.Errorf("create transcript directory: %w", err)
	}

	args := []string{
		"--input", audioPath,
		"--output", sourcePath,
		"--model", model,
		"--compute-type", computeType,
		"--language", language,
	}
	cmd := execCommand(ctx, "whisper-cli", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("whisper-cli failed: %w\n%s", err, string(output))
	}

	info, statErr := os.Stat(sourcePath)
	if statErr != nil {
		return fmt.Errorf("whisper-cli did not create source.vtt: %w", statErr)
	}
	if info.Size() == 0 {
		return fmt.Errorf("whisper-cli created empty source.vtt")
	}
	return nil
}
