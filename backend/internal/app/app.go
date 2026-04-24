package app

import (
	"net/http"
	"os"
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

	mockRunner := runner.NewMockRunner(database)
	handler := api.NewHandler(database, mockRunner, config.WorkDir)
	return api.Routes(handler), nil
}
