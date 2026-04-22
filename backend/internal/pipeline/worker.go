package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"lets-sub-it/backend/internal/jobs"
)

type JobService interface {
	GetJob(ctx context.Context, id string) (jobs.Job, error)
	ListJobsByStatus(ctx context.Context, status jobs.Status) ([]jobs.Job, error)
	ClaimQueuedJobs(ctx context.Context) ([]jobs.Job, error)
	UpdateProgress(ctx context.Context, jobID string, update jobs.ProgressUpdate) error
	SaveAsset(ctx context.Context, asset jobs.SubtitleAsset) error
}

type Worker struct {
	jobs        JobService
	downloader  Downloader
	transcriber Transcriber
	translator  Translator
	storageDir  string
}

func NewWorker(jobs JobService, downloader Downloader, transcriber Transcriber, translator Translator, storageDir string) *Worker {
	return &Worker{
		jobs:        jobs,
		downloader:  downloader,
		transcriber: transcriber,
		translator:  translator,
		storageDir:  storageDir,
	}
}

func (w *Worker) Run(ctx context.Context, jobID string) error {
	job, err := w.jobs.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if err := w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageDownloading,
		Progress: 10,
	}); err != nil {
		return err
	}

	mediaPath, err := w.downloader.Download(ctx, job.ID, job.YouTubeURL)
	if err != nil {
		return w.fail(ctx, jobID, jobs.StageDownloading, 10, err)
	}

	if err := w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageTranscribing,
		Progress: 35,
	}); err != nil {
		return err
	}

	segments, err := w.transcriber.Transcribe(ctx, mediaPath)
	if err != nil {
		return w.fail(ctx, jobID, jobs.StageTranscribing, 35, err)
	}

	if err := w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageTranslating,
		Progress: 70,
	}); err != nil {
		return err
	}

	translated, err := w.translator.TranslateSegments(ctx, job.TargetLanguage, segments)
	if err != nil {
		return w.fail(ctx, jobID, jobs.StageTranslating, 70, err)
	}

	translated, err = alignTranslatedSegments(segments, translated)
	if err != nil {
		return w.fail(ctx, jobID, jobs.StageTranslating, 70, err)
	}

	if err := w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StagePackaging,
		Progress: 90,
	}); err != nil {
		return err
	}

	sourcePath, translatedPath, bilingualPath, err := writeAssets(w.storageDir, job.VideoID, segments, translated)
	if err != nil {
		return w.fail(ctx, jobID, jobs.StagePackaging, 90, err)
	}

	if err := w.jobs.SaveAsset(ctx, jobs.SubtitleAsset{
		JobID:             job.ID,
		VideoID:           job.VideoID,
		SourceVTTPath:     sourcePath,
		TranslatedVTTPath: translatedPath,
		BilingualVTTPath:  bilingualPath,
		SourceLanguage:    "unknown",
		TargetLanguage:    job.TargetLanguage,
	}); err != nil {
		return w.fail(ctx, jobID, jobs.StagePackaging, 90, err)
	}

	return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusCompleted,
		Stage:    jobs.StageCompleted,
		Progress: 100,
	})
}

func (w *Worker) RunPendingLoop(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	if err := w.requeueInterruptedJobs(ctx); err != nil {
		return err
	}

	for {
		queuedJobs, err := w.jobs.ClaimQueuedJobs(ctx)
		if err != nil {
			return err
		}

		for _, job := range queuedJobs {
			if err := w.Run(ctx, job.ID); err != nil && ctx.Err() == nil {
				continue
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) requeueInterruptedJobs(ctx context.Context) error {
	runningJobs, err := w.jobs.ListJobsByStatus(ctx, jobs.StatusRunning)
	if err != nil {
		return err
	}

	for _, job := range runningJobs {
		if err := w.jobs.UpdateProgress(ctx, job.ID, jobs.ProgressUpdate{
			Status:   jobs.StatusQueued,
			Stage:    jobs.StageQueued,
			Progress: 0,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) fail(ctx context.Context, jobID string, stage jobs.Stage, progress int, err error) error {
	if updateErr := w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{
		Status:   jobs.StatusFailed,
		Stage:    stage,
		Progress: progress,
		Error:    err.Error(),
	}); updateErr != nil {
		return updateErr
	}

	return err
}

func writeAssets(storageDir string, videoID string, sourceSegments []Segment, translatedSegments []Segment) (string, string, string, error) {
	dir := filepath.Join(storageDir, videoID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", "", err
	}

	sourceVTT, err := BuildTranslatedVTT(sourceOnlySegments(sourceSegments))
	if err != nil {
		return "", "", "", err
	}

	translatedVTT, err := BuildTranslatedVTT(translatedSegments)
	if err != nil {
		return "", "", "", err
	}

	bilingualVTT, err := BuildBilingualVTT(translatedSegments)
	if err != nil {
		return "", "", "", err
	}

	sourcePath := filepath.Join(dir, "source.vtt")
	if err := os.WriteFile(sourcePath, []byte(sourceVTT), 0o644); err != nil {
		return "", "", "", err
	}

	translatedPath := filepath.Join(dir, "translated.vtt")
	if err := os.WriteFile(translatedPath, []byte(translatedVTT), 0o644); err != nil {
		return "", "", "", err
	}

	bilingualPath := filepath.Join(dir, "bilingual.vtt")
	if err := os.WriteFile(bilingualPath, []byte(bilingualVTT), 0o644); err != nil {
		return "", "", "", err
	}

	return sourcePath, translatedPath, bilingualPath, nil
}

func sourceOnlySegments(segments []Segment) []Segment {
	result := make([]Segment, len(segments))
	for i, segment := range segments {
		result[i] = Segment{
			Start:          segment.Start,
			End:            segment.End,
			TranslatedText: segment.SourceText,
		}
	}

	return result
}

func alignTranslatedSegments(sourceSegments []Segment, translatedSegments []Segment) ([]Segment, error) {
	if len(translatedSegments) != len(sourceSegments) {
		return nil, fmt.Errorf("translation returned %d segments for %d source segments", len(translatedSegments), len(sourceSegments))
	}

	result := make([]Segment, len(sourceSegments))
	for i, source := range sourceSegments {
		result[i] = Segment{
			Start:          source.Start,
			End:            source.End,
			SourceText:     source.SourceText,
			TranslatedText: translatedSegments[i].TranslatedText,
		}
	}

	return result, nil
}
