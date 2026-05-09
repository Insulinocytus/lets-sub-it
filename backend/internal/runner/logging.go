package runner

import (
	"log/slog"
	"time"

	"lets-sub-it-api/internal/store"
)

func logJobStarted(job store.Job) time.Time {
	startedAt := time.Now()
	slog.Info("job started",
		"job_id", job.ID,
		"video_id", job.VideoID,
		"source_language", job.SourceLanguage,
		"target_language", job.TargetLanguage,
	)
	return startedAt
}

func logJobCompleted(job store.Job, startedAt time.Time) {
	slog.Info("job completed",
		"job_id", job.ID,
		"video_id", job.VideoID,
		"duration_ms", time.Since(startedAt).Milliseconds(),
	)
}

func logJobFailed(job store.Job, stage string, cause error, startedAt time.Time) {
	slog.Error("job failed",
		"job_id", job.ID,
		"video_id", job.VideoID,
		"stage", stage,
		"duration_ms", time.Since(startedAt).Milliseconds(),
		"error", cause.Error(),
	)
}

func logJobStageStarted(job store.Job, stage string) time.Time {
	startedAt := time.Now()
	slog.Info("job stage started",
		"job_id", job.ID,
		"video_id", job.VideoID,
		"stage", stage,
	)
	return startedAt
}

func logJobStageCompleted(job store.Job, stage string, startedAt time.Time) {
	slog.Info("job stage completed",
		"job_id", job.ID,
		"video_id", job.VideoID,
		"stage", stage,
		"duration_ms", time.Since(startedAt).Milliseconds(),
	)
}
