package app

import (
	"fmt"
	"io"
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

// NewHTTPHandler returns the API handler and the database it holds open; the caller closes it.
func NewHTTPHandler(config Config) (http.Handler, io.Closer, error) {
	if err := checkTools(); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(filepath.Dir(config.DBPath), 0o755); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(config.WorkDir, 0o755); err != nil {
		return nil, nil, err
	}

	database, err := store.Open(config.DBPath)
	if err != nil {
		return nil, nil, err
	}
	if err := prepareDatabase(database); err != nil {
		_ = database.Close()
		return nil, nil, err
	}

	translator := runner.NewChatTranslator(config.LLMBaseURL, config.LLMAPIKey, config.LLMModel, config.LLMTimeout, http.DefaultClient)
	transcriber := runner.NewHTTPTranscriber(config.STTBaseURL, config.STTAPIKey, config.STTTimeout, http.DefaultClient)
	jobRunner := runner.NewRealRunner(database, config.DownloadTimeout, config.STTModel, transcriber, translator)

	handler := api.NewHandler(database, jobRunner, config.WorkDir)
	return api.Routes(handler), database, nil
}

func prepareDatabase(database *store.Store) error {
	if err := database.Migrate(); err != nil {
		return err
	}
	interruptedCount, err := database.FailInterruptedJobs("任务因后端重启中断，请重新提交")
	if err != nil {
		return err
	}
	if interruptedCount > 0 {
		slog.Warn("interrupted jobs marked failed", "count", interruptedCount)
	}
	return nil
}

func checkTools() error {
	for _, tool := range []string{"yt-dlp", "ffmpeg"} {
		if _, err := lookPath(tool); err != nil {
			return fmt.Errorf("backend requires %s to be installed and on PATH", tool)
		}
	}
	return nil
}
