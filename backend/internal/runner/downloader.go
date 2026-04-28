package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var execCommand = exec.CommandContext

func downloadAudio(ctx context.Context, workDir string, jobID string, youtubeURL string) (string, error) {
	jobDir := filepath.Join(workDir, jobID)
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		return "", fmt.Errorf("create job directory: %w", err)
	}
	audioPath := filepath.Join(jobDir, "audio.mp3")
	args := []string{
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "128K",
		"-o", filepath.Join(jobDir, "audio.%(ext)s"),
		youtubeURL,
	}
	cmd := execCommand(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w\n%s", err, string(output))
	}
	return audioPath, nil
}
