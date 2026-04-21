package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Downloader interface {
	Download(ctx context.Context, jobID string, youtubeURL string) (string, error)
}

type fakeDownloader struct {
	path string
}

func FakeDownloader(path string) Downloader {
	return fakeDownloader{path: path}
}

func (f fakeDownloader) Download(_ context.Context, _ string, _ string) (string, error) {
	return f.path, nil
}

type ytDLPDownloader struct {
	binary string
}

func NewYTDLPDownloader() Downloader {
	return ytDLPDownloader{binary: "yt-dlp"}
}

func (d ytDLPDownloader) Download(ctx context.Context, jobID string, youtubeURL string) (string, error) {
	outputDir, err := os.MkdirTemp("", "lets-sub-it-download-*")
	if err != nil {
		return "", err
	}

	outputTemplate := filepath.Join(outputDir, jobID+".%(ext)s")
	cmd := exec.CommandContext(
		ctx,
		d.binary,
		"--no-playlist",
		"-f", "bestaudio/best",
		"-o", outputTemplate,
		youtubeURL,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w: %s", err, string(output))
	}

	matches, err := filepath.Glob(filepath.Join(outputDir, jobID+".*"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("yt-dlp did not produce an output file for job %s", jobID)
	}

	return matches[0], nil
}
