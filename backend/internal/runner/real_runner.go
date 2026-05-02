package runner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"lets-sub-it-api/internal/store"
)

type RealRunner struct {
	store              Store
	downloadTimeout    time.Duration
	whisperModel       string
	whisperComputeType string
	translator         Translator
}

func NewRealRunner(store Store, downloadTimeout time.Duration, whisperModel string, whisperComputeType string, translator Translator) *RealRunner {
	return &RealRunner{store: store, downloadTimeout: downloadTimeout, whisperModel: whisperModel, whisperComputeType: whisperComputeType, translator: translator}
}

func (r *RealRunner) Start(ctx context.Context, job store.Job) error {
	if err := r.set(job.ID, store.StatusDownloading, "正在下载音频...", ""); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}
	if err := ctx.Err(); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}

	downloadCtx, cancel := context.WithTimeout(ctx, r.downloadTimeout)
	defer cancel()

	// WorkingDir format: <LSI_WORK_DIR>/<jobID> (see store.NewJob)
	audioPath, err := downloadAudio(downloadCtx, filepath.Dir(job.WorkingDir), job.ID, job.YoutubeURL)
	if err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}

	if err := r.set(job.ID, store.StatusTranscribing, "调用 whisper-cli 生成 source.vtt", ""); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}
	if err := os.MkdirAll(job.WorkingDir, 0o755); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	sourcePath := filepath.Join(job.WorkingDir, "source.vtt")
	translatedPath := filepath.Join(job.WorkingDir, "translated.vtt")
	bilingualPath := filepath.Join(job.WorkingDir, "bilingual.vtt")

	if err := transcribeAudio(ctx, audioPath, sourcePath, r.whisperModel, r.whisperComputeType, job.SourceLanguage); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	if err := r.set(job.ID, store.StatusTranslating, "翻译字幕...", ""); err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}
	cues, err := parseWebVTTCues(string(sourceContent))
	if err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}
	translations, err := r.translator.Translate(ctx, cues, job.SourceLanguage, job.TargetLanguage)
	if err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}
	translatedVTT, err := renderTranslatedVTT(cues, translations)
	if err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}

	if err := os.WriteFile(translatedPath, []byte(translatedVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}

	if err := r.set(job.ID, store.StatusPackaging, "生成字幕资产", ""); err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
	}
	bilingualVTT, err := renderBilingualVTT(cues, translations)
	if err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
	}
	if err := os.WriteFile(bilingualPath, []byte(bilingualVTT), 0o644); err != nil {
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

	if err := r.set(job.ID, store.StatusCompleted, "处理完成", ""); err != nil {
		return r.fail(job.ID, store.StatusCompleted, err)
	}
	return nil
}

func (r *RealRunner) set(jobID string, status string, progressText string, errorMessage string) error {
	return r.store.UpdateJobStatus(jobID, status, status, progressText, errorMessage)
}

func (r *RealRunner) fail(jobID string, stage string, cause error) error {
	if updateErr := r.store.UpdateJobStatus(jobID, store.StatusFailed, stage, "处理失败", cause.Error()); updateErr != nil {
		return updateErr
	}
	return cause
}
