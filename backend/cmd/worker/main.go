package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lets-sub-it/backend/internal/config"
	"lets-sub-it/backend/internal/db"
	"lets-sub-it/backend/internal/jobs"
	"lets-sub-it/backend/internal/pipeline"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, config.Load(), log.Default()); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg config.Config, logger *log.Logger) error {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	if err := os.MkdirAll(cfg.StorageDir, 0o755); err != nil {
		return err
	}

	sqlDB, err := db.Open(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	repo, err := jobs.NewSQLiteRepository(sqlDB)
	if err != nil {
		return err
	}

	service := jobs.NewService(repo, time.Now)
	worker := pipeline.NewWorker(
		service,
		pipeline.NewYTDLPDownloader(),
		pipeline.NewFastWhisperTranscriber(),
		pipeline.NewOpenAITranslator(),
		cfg.StorageDir,
	)

	logger.Printf("worker polling queued jobs using database %s", cfg.DatabasePath)
	return worker.RunPendingLoop(ctx)
}
