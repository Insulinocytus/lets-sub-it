package runner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"lets-sub-it-api/internal/store"
)

type MockRunner struct {
	store Store
}

func NewMockRunner(store Store) *MockRunner {
	return &MockRunner{store: store}
}

func (r *MockRunner) Start(ctx context.Context, job store.Job) error {
	jobStartedAt := logJobStarted(job, "mock")
	stageStartedAt := logJobStageStarted(job, store.StatusDownloading)
	if err := r.set(job.ID, store.StatusDownloading, "准备 mock 媒体", ""); err != nil {
		return r.fail(job, store.StatusDownloading, err, jobStartedAt)
	}
	if err := ctx.Err(); err != nil {
		return r.fail(job, store.StatusDownloading, err, jobStartedAt)
	}
	logJobStageCompleted(job, store.StatusDownloading, stageStartedAt)

	stageStartedAt = logJobStageStarted(job, store.StatusTranscribing)
	if err := r.set(job.ID, store.StatusTranscribing, "生成 mock source.vtt", ""); err != nil {
		return r.fail(job, store.StatusTranscribing, err, jobStartedAt)
	}
	if err := os.MkdirAll(job.WorkingDir, 0o755); err != nil {
		return r.fail(job, store.StatusTranscribing, err, jobStartedAt)
	}

	sourcePath := filepath.Join(job.WorkingDir, "source.vtt")
	translatedPath := filepath.Join(job.WorkingDir, "translated.vtt")
	bilingualPath := filepath.Join(job.WorkingDir, "bilingual.vtt")

	if err := os.WriteFile(sourcePath, []byte(mockSourceVTT), 0o644); err != nil {
		return r.fail(job, store.StatusTranscribing, err, jobStartedAt)
	}
	logJobStageCompleted(job, store.StatusTranscribing, stageStartedAt)

	stageStartedAt = logJobStageStarted(job, store.StatusTranslating)
	for _, progress := range []string{"1/3 segments", "2/3 segments", "3/3 segments"} {
		if err := r.set(job.ID, store.StatusTranslating, progress, ""); err != nil {
			return r.fail(job, store.StatusTranslating, err, jobStartedAt)
		}
	}

	if err := os.WriteFile(translatedPath, []byte(mockTranslatedVTT), 0o644); err != nil {
		return r.fail(job, store.StatusTranslating, err, jobStartedAt)
	}
	logJobStageCompleted(job, store.StatusTranslating, stageStartedAt)

	stageStartedAt = logJobStageStarted(job, store.StatusPackaging)
	if err := r.set(job.ID, store.StatusPackaging, "生成字幕资产", ""); err != nil {
		return r.fail(job, store.StatusPackaging, err, jobStartedAt)
	}
	if err := os.WriteFile(bilingualPath, []byte(mockBilingualVTT), 0o644); err != nil {
		return r.fail(job, store.StatusPackaging, err, jobStartedAt)
	}

	asset := store.SubtitleAsset{
		JobID:             job.ID,
		VideoID:           job.VideoID,
		TargetLanguage:    job.TargetLanguage,
		SourceLanguage:    job.SourceLanguage,
		SourceVTTPath:     sourcePath,
		TranslatedVTTPath: translatedPath,
		BilingualVTTPath:  bilingualPath,
	}
	if err := r.store.CreateSubtitleAsset(asset); err != nil {
		return r.fail(job, store.StatusPackaging, err, jobStartedAt)
	}
	logJobStageCompleted(job, store.StatusPackaging, stageStartedAt)

	if err := r.set(job.ID, store.StatusCompleted, "处理完成", ""); err != nil {
		return r.fail(job, store.StatusCompleted, err, jobStartedAt)
	}
	logJobCompleted(job, jobStartedAt)
	return nil
}

func (r *MockRunner) set(jobID string, status string, progressText string, errorMessage string) error {
	return r.store.UpdateJobStatus(jobID, status, status, progressText, errorMessage)
}

func (r *MockRunner) fail(job store.Job, stage string, cause error, jobStartedAt time.Time) error {
	logJobFailed(job, stage, cause, jobStartedAt)
	if updateErr := r.store.UpdateJobStatus(job.ID, store.StatusFailed, stage, "处理失败", cause.Error()); updateErr != nil {
		return updateErr
	}
	return cause
}
