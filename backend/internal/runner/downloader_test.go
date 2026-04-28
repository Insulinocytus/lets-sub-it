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
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	// Intentionally do NOT create the job directory — downloadAudio should create it.

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

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
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	jobDir := filepath.Join(tmpDir, "job_1")

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo fake-audio-data > "+filepath.Join(jobDir, "audio.mp3"))
	}

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
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Video unavailable' >&2 && exit 1")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=deleted")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Video unavailable") {
		t.Fatalf("error = %q, want containing 'Video unavailable'", err.Error())
	}
}

func TestDownloadAudioTimeout(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "10")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := downloadAudio(ctx, tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want context deadline exceeded")
	}
}

func TestDownloadAudioNetworkError(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Unable to download webpage: network error' >&2 && exit 1")
	}

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
