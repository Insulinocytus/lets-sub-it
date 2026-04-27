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

func TestDownloadAudioSuccess(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	jobDir := filepath.Join(tmpDir, "job_1")
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "mkdir -p "+jobDir+" && echo fake-audio-data > "+jobDir+"/audio.mp3")
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
	jobDir := filepath.Join(tmpDir, "job_1")
	os.MkdirAll(jobDir, 0o755)

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
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

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
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

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
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "yt-dlp-this-tool-does-not-exist-xyz")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want exec error")
	}
}
