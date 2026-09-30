package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lets-sub-it-api/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	testStore, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = testStore.Close() })
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return testStore
}

// Fake yt-dlp behaviours. A mode of "fail:<message>" prints message to stderr and exits 1.
const (
	ytDlpWritesAudio  = "writes-audio"
	ytDlpExitsCleanly = "exits-cleanly"
	ytDlpHangs        = "hangs"
)

// useFakeYtDlp routes execCommand to this test binary acting as yt-dlp (see TestHelperProcess),
// so the tests need no POSIX shell.
func useFakeYtDlp(t *testing.T, mode string) {
	t.Helper()
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestHelperProcess$", "--", name}, args...)...)
		cmd.Env = append(os.Environ(), "LSI_FAKE_YTDLP="+mode)
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	mode := os.Getenv("LSI_FAKE_YTDLP")
	if mode == "" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 || args[0] != "yt-dlp" {
		fmt.Fprintf(os.Stderr, "unexpected command %q\n", args)
		os.Exit(127)
	}

	switch {
	case strings.HasPrefix(mode, "fail:"):
		fmt.Fprintln(os.Stderr, strings.TrimPrefix(mode, "fail:"))
		os.Exit(1)
	case mode == ytDlpHangs:
		time.Sleep(time.Minute)
	case mode == ytDlpWritesAudio:
		for i := 1; i+1 < len(args); i++ {
			if args[i] != "-o" {
				continue
			}
			path := strings.Replace(args[i+1], "%(ext)s", "mp3", 1)
			if err := os.WriteFile(path, []byte("fake-audio-data"), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
	os.Exit(0)
}
