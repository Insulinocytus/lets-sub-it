package main

import (
	"log/slog"
	"net/http"
	"os"

	"lets-sub-it-api/internal/app"
)

func main() {
	config := app.LoadConfig()
	app.ConfigureLogger(config.LogLevel, os.Stdout)

	handler, err := app.NewHTTPHandler(config)
	if err != nil {
		slog.Error("server initialization failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server starting", "addr", config.Addr, "runner_mode", config.RunnerMode, "log_level", config.LogLevel)
	if err := http.ListenAndServe(config.Addr, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
