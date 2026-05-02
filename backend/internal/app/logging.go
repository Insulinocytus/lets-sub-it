package app

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func ConfigureLogger(levelName string, writer io.Writer) func() {
	if writer == nil {
		writer = os.Stdout
	}

	previous := slog.Default()
	level := slogLevel(levelName)
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})))
	return func() {
		slog.SetDefault(previous)
	}
}

func slogLevel(levelName string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelName)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
