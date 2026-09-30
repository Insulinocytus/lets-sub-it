package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadAudioCreatesJobDir(t *testing.T) {
	useFakeYtDlp(t, ytDlpExitsCleanly)

	tmpDir := t.TempDir()
	// Intentionally do NOT create the job directory — downloadAudio should create it.

	_, err := downloadAudio(context.Background(), tmpDir, "job_newdir", "https://www.youtube.com/watch?v=abc123")
	if err != nil {
		t.Fatalf("downloadAudio() error = %v", err)
	}

	jobDir := filepath.Join(tmpDir, "job_newdir")
	info, statErr := os.Stat(jobDir)
	if statErr != nil {
		t.Fatalf("os.Stat(jobDir) error = %v", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("jobDir is not a directory")
	}
}

func TestDownloadAudioSuccess(t *testing.T) {
	useFakeYtDlp(t, ytDlpWritesAudio)

	tmpDir := t.TempDir()
	jobDir := filepath.Join(tmpDir, "job_1")

	audioPath, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err != nil {
		t.Fatalf("downloadAudio() error = %v", err)
	}
	if audioPath != filepath.Join(jobDir, "audio.mp3") {
		t.Fatalf("audioPath = %q, want %q", audioPath, filepath.Join(jobDir, "audio.mp3"))
	}

	data, readErr := os.ReadFile(audioPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(audio.mp3) error = %v", readErr)
	}
	if len(data) == 0 {
		t.Fatal("audio.mp3 is empty")
	}
}

func TestDownloadAudioVideoUnavailable(t *testing.T) {
	useFakeYtDlp(t, "fail:ERROR: Video unavailable")

	tmpDir := t.TempDir()

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=deleted")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Video unavailable") {
		t.Fatalf("error = %q, want containing 'Video unavailable'", err.Error())
	}
}

func TestDownloadAudioTimeout(t *testing.T) {
	useFakeYtDlp(t, ytDlpHangs)

	tmpDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := downloadAudio(ctx, tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want context deadline exceeded")
	}
}

func TestDownloadAudioNetworkError(t *testing.T) {
	useFakeYtDlp(t, "fail:ERROR: Unable to download webpage: network error")

	tmpDir := t.TempDir()

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Fatalf("error = %q, want containing 'network error'", err.Error())
	}
}

func TestDownloadAudioYtDlpMissing(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "yt-dlp-this-tool-does-not-exist-xyz")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want exec error")
	}
}
