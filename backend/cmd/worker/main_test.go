package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"lets-sub-it/backend/internal/config"
)

func TestRunInitializesStorageDirAndStopsOnCancel(t *testing.T) {
	storageDir := filepath.Join(t.TempDir(), "assets")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, config.Config{StorageDir: storageDir}, log.New(io.Discard, "", 0))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	info, statErr := os.Stat(storageDir)
	if statErr != nil {
		t.Fatalf("expected storage dir to exist, got %v", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", storageDir)
	}
}
