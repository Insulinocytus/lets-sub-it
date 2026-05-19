package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"lets-sub-it-api/internal/api"
	"lets-sub-it-api/internal/runner"
	"lets-sub-it-api/internal/store"
)

var lookPath = exec.LookPath

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
	interruptedCount, err := database.FailInterruptedJobs("任务因后端重启中断，请重新提交")
	if err != nil {
		return nil, err
	}
	if interruptedCount > 0 {
		slog.Warn("interrupted jobs marked failed", "count", interruptedCount)
	}

	if err := checkTools(); err != nil {
		return nil, err
	}
	translator := runner.NewChatTranslator(config.LLMBaseURL, config.LLMAPIKey, config.LLMModel, config.LLMTimeout, http.DefaultClient)
	transcriber := runner.NewHTTPTranscriber(config.WhisperBaseURL, config.WhisperTimeout, config.WhisperPollInterval, http.DefaultClient)
	jobRunner := runner.NewRealRunner(database, config.DownloadTimeout, config.WhisperModel, config.WhisperComputeType, transcriber, translator)

	handler := api.NewHandler(database, jobRunner, config.WorkDir)
	return api.Routes(handler), nil
}

func checkTools() error {
	for _, tool := range []string{"yt-dlp", "ffmpeg"} {
		if _, err := lookPath(tool); err != nil {
			return fmt.Errorf("backend requires %s to be installed and on PATH", tool)
		}
	}
	return nil
}
