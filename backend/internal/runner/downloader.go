package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
	startedAt := time.Now()
	slog.Debug("external command started", "job_id", jobID, "command", "yt-dlp")
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug("external command failed", "job_id", jobID, "command", "yt-dlp", "duration_ms", time.Since(startedAt).Milliseconds())
		return "", fmt.Errorf("yt-dlp failed: %w\n%s", err, string(output))
	}
	slog.Debug("external command completed", "job_id", jobID, "command", "yt-dlp", "duration_ms", time.Since(startedAt).Milliseconds())
	return audioPath, nil
}
