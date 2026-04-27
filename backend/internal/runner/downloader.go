package runner

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

var execCommand = exec.CommandContext

func downloadAudio(ctx context.Context, workDir string, jobID string, youtubeURL string) (string, error) {
	audioPath := filepath.Join(workDir, jobID, "audio.mp3")
	args := []string{
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "128K",
		"-o", filepath.Join(workDir, jobID, "audio.%(ext)s"),
		youtubeURL,
	}
	cmd := execCommand(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w\n%s", err, string(output))
	}
	return audioPath, nil
}
