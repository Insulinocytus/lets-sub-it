package app

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestConfigureLoggerFiltersByLevel(t *testing.T) {
	var output bytes.Buffer
	restore := ConfigureLogger("warn", &output)
	defer restore()

	slog.Debug("debug message")
	slog.Info("info message")
	slog.Warn("warn message")

	logs := output.String()
	if strings.Contains(logs, "debug message") {
		t.Fatalf("logs = %q, want no debug message", logs)
	}
	if strings.Contains(logs, "info message") {
		t.Fatalf("logs = %q, want no info message", logs)
	}
	if !strings.Contains(logs, "warn message") {
		t.Fatalf("logs = %q, want warn message", logs)
	}
}

func TestConfigureLoggerDefaultsInvalidLevelToInfo(t *testing.T) {
	var output bytes.Buffer
	restore := ConfigureLogger("not-a-level", &output)
	defer restore()

	slog.Debug("debug message")
	slog.Info("info message")

	logs := output.String()
	if strings.Contains(logs, "debug message") {
		t.Fatalf("logs = %q, want no debug message", logs)
	}
	if !strings.Contains(logs, "info message") {
		t.Fatalf("logs = %q, want info message", logs)
	}
}
