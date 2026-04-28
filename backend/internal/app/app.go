package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"lets-sub-it-api/internal/api"
	"lets-sub-it-api/internal/runner"
	"lets-sub-it-api/internal/store"
)

func NewHTTPHandler(config Config) (http.Handler, error) {
	if err := os.MkdirAll(filepath.Dir(config.DBPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.WorkDir, 0o755); err != nil {
		return nil, err
	}

	database, err := store.Open(config.DBPath)
	if err != nil {
		return nil, err
	}
	if err := database.Migrate(); err != nil {
		return nil, err
	}

	var jobRunner runner.Runner
	switch config.RunnerMode {
	case "real":
		if err := checkTools(); err != nil {
			return nil, err
		}
		jobRunner = runner.NewRealRunner(database, config.DownloadTimeout)
	default:
		jobRunner = runner.NewMockRunner(database)
	}

	handler := api.NewHandler(database, jobRunner, config.WorkDir)
	return api.Routes(handler), nil
}

func checkTools() error {
	for _, tool := range []string{"yt-dlp", "ffmpeg"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("LSI_RUNNER_MODE=real requires %s to be installed and on PATH", tool)
		}
	}
	return nil
}
