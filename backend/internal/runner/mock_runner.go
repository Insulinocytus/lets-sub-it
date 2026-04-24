package runner

import (
	"context"
	"os"
	"path/filepath"

	"lets-sub-it-api/internal/store"
)

type MockRunner struct {
	store Store
}

func NewMockRunner(store Store) *MockRunner {
	return &MockRunner{store: store}
}

func (r *MockRunner) Start(ctx context.Context, job store.Job) error {
	if err := r.set(job.ID, store.StatusDownloading, "准备 mock 媒体", ""); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}

	if err := r.set(job.ID, store.StatusTranscribing, "生成 mock source.vtt", ""); err != nil {
		return err
	}
	if err := os.MkdirAll(job.WorkingDir, 0o755); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	sourcePath := filepath.Join(job.WorkingDir, "source.vtt")
	translatedPath := filepath.Join(job.WorkingDir, "translated.vtt")
	bilingualPath := filepath.Join(job.WorkingDir, "bilingual.vtt")

	if err := os.WriteFile(sourcePath, []byte(mockSourceVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	for _, progress := range []string{"1/3 segments", "2/3 segments", "3/3 segments"} {
		if err := r.set(job.ID, store.StatusTranslating, progress, ""); err != nil {
			return err
		}
	}

	if err := os.WriteFile(translatedPath, []byte(mockTranslatedVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}

	if err := r.set(job.ID, store.StatusPackaging, "生成字幕资产", ""); err != nil {
		return err
	}
	if err := os.WriteFile(bilingualPath, []byte(mockBilingualVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
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
		return r.fail(job.ID, store.StatusPackaging, err)
	}

	return r.set(job.ID, store.StatusCompleted, "处理完成", "")
}

func (r *MockRunner) set(jobID string, status string, progressText string, errorMessage string) error {
	return r.store.UpdateJobStatus(jobID, status, status, progressText, errorMessage)
}

func (r *MockRunner) fail(jobID string, stage string, cause error) error {
	if updateErr := r.store.UpdateJobStatus(jobID, store.StatusFailed, stage, "处理失败", cause.Error()); updateErr != nil {
		return updateErr
	}
	return cause
}
