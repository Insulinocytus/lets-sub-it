package jobs_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"lets-sub-it/backend/internal/db"
	"lets-sub-it/backend/internal/jobs"
	"lets-sub-it/backend/internal/pipeline"
)

type failingDownloader struct {
	err error
}

func (d failingDownloader) Download(context.Context, string, string) (string, error) {
	return "", d.err
}

type saveAssetFailingRepository struct {
	jobs.Repository
	err error
}

func (r saveAssetFailingRepository) SaveAsset(context.Context, jobs.SubtitleAsset) error {
	return r.err
}

func TestServiceCreateAndUpdateJob(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	createTime := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	updateTime := createTime.Add(2 * time.Minute)
	nowCalls := 0
	svc := jobs.NewService(repo, func() time.Time {
		nowCalls++
		if nowCalls == 1 {
			return createTime
		}
		return updateTime
	})

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if job.Status != jobs.StatusQueued {
		t.Fatalf("expected status %q, got %q", jobs.StatusQueued, job.Status)
	}
	if job.Stage != jobs.StageQueued {
		t.Fatalf("expected stage %q, got %q", jobs.StageQueued, job.Stage)
	}
	if job.ID != fmt.Sprintf("job-%d", createTime.UnixNano()) {
		t.Fatalf("expected deterministic id, got %q", job.ID)
	}
	if !job.CreatedAt.Equal(createTime) {
		t.Fatalf("expected CreatedAt %v, got %v", createTime, job.CreatedAt)
	}
	if !job.UpdatedAt.Equal(createTime) {
		t.Fatalf("expected UpdatedAt %v, got %v", createTime, job.UpdatedAt)
	}

	err = svc.UpdateProgress(context.Background(), job.ID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageDownloading,
		Progress: 42,
	})
	if err != nil {
		t.Fatalf("UpdateProgress returned error: %v", err)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}

	if stored.Stage != jobs.StageDownloading {
		t.Fatalf("expected stage %q, got %q", jobs.StageDownloading, stored.Stage)
	}
	if stored.Progress != 42 {
		t.Fatalf("expected progress 42, got %d", stored.Progress)
	}
	if !stored.CreatedAt.Equal(createTime) {
		t.Fatalf("expected CreatedAt %v, got %v", createTime, stored.CreatedAt)
	}
	if !stored.UpdatedAt.Equal(updateTime) {
		t.Fatalf("expected UpdatedAt %v, got %v", updateTime, stored.UpdatedAt)
	}
}

func TestServiceCreateJobValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input jobs.CreateJobInput
	}{
		{
			name: "missing video id",
			input: jobs.CreateJobInput{
				YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
				TargetLanguage: "zh-CN",
			},
		},
		{
			name: "missing youtube url",
			input: jobs.CreateJobInput{
				VideoID:        "abc123xyz00",
				TargetLanguage: "zh-CN",
			},
		},
		{
			name: "missing target language",
			input: jobs.CreateJobInput{
				VideoID:    "abc123xyz00",
				YouTubeURL: "https://www.youtube.com/watch?v=abc123xyz00",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := jobs.NewMemoryRepositoryForTest()
			svc := jobs.NewService(repo, func() time.Time { return time.Unix(0, 1) })

			_, err := svc.CreateJob(context.Background(), tc.input)
			if !errors.Is(err, jobs.ErrInvalidCreateJobInput) {
				t.Fatalf("expected ErrInvalidCreateJobInput, got %v", err)
			}
		})
	}
}

func TestServiceUpdateProgressValidation(t *testing.T) {
	testCases := []struct {
		name   string
		jobID  string
		update jobs.ProgressUpdate
	}{
		{
			name:  "missing job id",
			jobID: "",
			update: jobs.ProgressUpdate{
				Status:   jobs.StatusRunning,
				Stage:    jobs.StageDownloading,
				Progress: 10,
			},
		},
		{
			name:  "missing status",
			jobID: "job-1",
			update: jobs.ProgressUpdate{
				Stage:    jobs.StageDownloading,
				Progress: 10,
			},
		},
		{
			name:  "missing stage",
			jobID: "job-1",
			update: jobs.ProgressUpdate{
				Status:   jobs.StatusRunning,
				Progress: 10,
			},
		},
		{
			name:  "progress below range",
			jobID: "job-1",
			update: jobs.ProgressUpdate{
				Status:   jobs.StatusRunning,
				Stage:    jobs.StageDownloading,
				Progress: -1,
			},
		},
		{
			name:  "progress above range",
			jobID: "job-1",
			update: jobs.ProgressUpdate{
				Status:   jobs.StatusRunning,
				Stage:    jobs.StageDownloading,
				Progress: 101,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := jobs.NewMemoryRepositoryForTest()
			svc := jobs.NewService(repo, func() time.Time { return time.Unix(0, 1) })

			err := svc.UpdateProgress(context.Background(), tc.jobID, tc.update)
			if !errors.Is(err, jobs.ErrInvalidProgressUpdate) {
				t.Fatalf("expected ErrInvalidProgressUpdate, got %v", err)
			}
		})
	}
}

func TestServiceUpdateProgressMissingJob(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	svc := jobs.NewService(repo, func() time.Time { return time.Unix(0, 1) })

	err := svc.UpdateProgress(context.Background(), "job-missing", jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageDownloading,
		Progress: 10,
	})
	if !errors.Is(err, jobs.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestServiceCreateJobDuplicateID(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })

	firstJob, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("first CreateJob returned error: %v", err)
	}

	_, err = svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "def456uvw99",
		YouTubeURL:     "https://www.youtube.com/watch?v=def456uvw99",
		TargetLanguage: "ja",
	})
	if !errors.Is(err, jobs.ErrJobAlreadyExists) {
		t.Fatalf("expected ErrJobAlreadyExists, got %v", err)
	}

	stored, err := svc.GetJob(context.Background(), firstJob.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}
	if stored.VideoID != firstJob.VideoID {
		t.Fatalf("expected original job to remain stored, got video id %q", stored.VideoID)
	}
}

func TestInMemoryRepositoryReturnsErrJobNotFound(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()

	_, err := repo.GetJob(context.Background(), "job-missing")
	if !errors.Is(err, jobs.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound from GetJob, got %v", err)
	}

	err = repo.UpdateJob(context.Background(), jobs.Job{ID: "job-missing"})
	if !errors.Is(err, jobs.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound from UpdateJob, got %v", err)
	}
}

func TestInMemoryRepositoryInsertJobReturnsErrJobAlreadyExists(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	job := jobs.Job{ID: "job-1", VideoID: "abc123xyz00"}

	if err := repo.InsertJob(context.Background(), job); err != nil {
		t.Fatalf("first InsertJob returned error: %v", err)
	}

	err := repo.InsertJob(context.Background(), jobs.Job{ID: "job-1", VideoID: "def456uvw99"})
	if !errors.Is(err, jobs.ErrJobAlreadyExists) {
		t.Fatalf("expected ErrJobAlreadyExists, got %v", err)
	}

	stored, err := repo.GetJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}
	if stored.VideoID != job.VideoID {
		t.Fatalf("expected original job to remain stored, got video id %q", stored.VideoID)
	}
}

func TestWorkerRunCompletesJobAndStoresAssets(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	nowCalls := 0
	svc := jobs.NewService(repo, func() time.Time {
		nowCalls++
		return now.Add(time.Duration(nowCalls-1) * time.Nanosecond)
	})
	storageDir := t.TempDir()
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		storageDir,
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if err := worker.Run(context.Background(), job.ID); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}

	if stored.Status != jobs.StatusCompleted {
		t.Fatalf("expected completed status, got %q", stored.Status)
	}
	if stored.Stage != jobs.StageCompleted {
		t.Fatalf("expected completed stage, got %q", stored.Stage)
	}
	if stored.Progress != 100 {
		t.Fatalf("expected progress 100, got %d", stored.Progress)
	}

	asset, err := svc.GetAssetByVideoID(context.Background(), job.VideoID)
	if err != nil {
		t.Fatalf("GetAssetByVideoID returned error: %v", err)
	}

	if asset.JobID != job.ID {
		t.Fatalf("expected asset job id %q, got %q", job.ID, asset.JobID)
	}
	if asset.TargetLanguage != "zh-CN" {
		t.Fatalf("expected target language zh-CN, got %q", asset.TargetLanguage)
	}
	if asset.SourceLanguage != "unknown" {
		t.Fatalf("expected source language unknown, got %q", asset.SourceLanguage)
	}

	expectedRelativeDir := filepath.ToSlash(filepath.Join(job.VideoID, job.ID))
	if asset.SourceVTTPath != expectedRelativeDir+"/source.vtt" {
		t.Fatalf("unexpected source path %q", asset.SourceVTTPath)
	}
	if asset.TranslatedVTTPath != expectedRelativeDir+"/translated.vtt" {
		t.Fatalf("unexpected translated path %q", asset.TranslatedVTTPath)
	}
	if asset.BilingualVTTPath != expectedRelativeDir+"/bilingual.vtt" {
		t.Fatalf("unexpected bilingual path %q", asset.BilingualVTTPath)
	}

	sourceBody, err := os.ReadFile(filepath.Join(storageDir, asset.SourceVTTPath))
	if err != nil {
		t.Fatalf("ReadFile(source) returned error: %v", err)
	}
	if !strings.Contains(string(sourceBody), "Hello") {
		t.Fatalf("expected source VTT to contain source text, got %q", string(sourceBody))
	}

	translatedBody, err := os.ReadFile(filepath.Join(storageDir, asset.TranslatedVTTPath))
	if err != nil {
		t.Fatalf("ReadFile(translated) returned error: %v", err)
	}
	if !strings.Contains(string(translatedBody), "你好") {
		t.Fatalf("expected translated VTT to contain translated text, got %q", string(translatedBody))
	}

	bilingualBody, err := os.ReadFile(filepath.Join(storageDir, asset.BilingualVTTPath))
	if err != nil {
		t.Fatalf("ReadFile(bilingual) returned error: %v", err)
	}
	if !strings.Contains(string(bilingualBody), "Hello") || !strings.Contains(string(bilingualBody), "你好") {
		t.Fatalf("expected bilingual VTT to contain source and translated text, got %q", string(bilingualBody))
	}
}

func TestWorkerRunMarksJobFailedWhenDownloadFails(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	nowCalls := 0
	svc := jobs.NewService(repo, func() time.Time {
		nowCalls++
		return now.Add(time.Duration(nowCalls-1) * time.Nanosecond)
	})
	worker := pipeline.NewWorker(
		svc,
		failingDownloader{err: errors.New("download failed")},
		pipeline.FakeTranscriber(nil),
		pipeline.FakeTranslator(nil),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	err = worker.Run(context.Background(), job.ID)
	if err == nil || err.Error() != "download failed" {
		t.Fatalf("expected original download error, got %v", err)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}

	if stored.Status != jobs.StatusFailed {
		t.Fatalf("expected failed status, got %q", stored.Status)
	}
	if stored.Stage != jobs.StageDownloading {
		t.Fatalf("expected downloading stage, got %q", stored.Stage)
	}
	if stored.Progress != 10 {
		t.Fatalf("expected progress 10, got %d", stored.Progress)
	}
	if stored.ErrorMessage != "download failed" {
		t.Fatalf("expected error message to be stored, got %q", stored.ErrorMessage)
	}
}

func TestWorkerRunMarksJobFailedWhenTranslationSegmentCountDoesNotAlign(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{
			{
				Start:      "00:00:01.000",
				End:        "00:00:03.000",
				SourceText: "Hello",
			},
			{
				Start:      "00:00:03.000",
				End:        "00:00:05.000",
				SourceText: "World",
			},
		}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	err = worker.Run(context.Background(), job.ID)
	if err == nil {
		t.Fatal("expected Run to fail for misaligned translation output")
	}
	if !strings.Contains(err.Error(), "translation returned 1 segments for 2 source segments") {
		t.Fatalf("unexpected Run error: %v", err)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}

	if stored.Status != jobs.StatusFailed {
		t.Fatalf("expected failed status, got %q", stored.Status)
	}
	if stored.Stage != jobs.StageTranslating {
		t.Fatalf("expected translating stage, got %q", stored.Stage)
	}
	if stored.Progress != 70 {
		t.Fatalf("expected progress 70, got %d", stored.Progress)
	}
	if !strings.Contains(stored.ErrorMessage, "translation returned 1 segments for 2 source segments") {
		t.Fatalf("expected alignment error to be stored, got %q", stored.ErrorMessage)
	}
}

func TestWorkerRunMarksJobFailedWhenSaveAssetFails(t *testing.T) {
	repo := saveAssetFailingRepository{
		Repository: jobs.NewMemoryRepositoryForTest(),
		err:        errors.New("save asset failed"),
	}
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	err = worker.Run(context.Background(), job.ID)
	if err == nil || err.Error() != "save asset failed" {
		t.Fatalf("expected original save asset error, got %v", err)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}

	if stored.Status != jobs.StatusFailed {
		t.Fatalf("expected failed status, got %q", stored.Status)
	}
	if stored.Stage != jobs.StagePackaging {
		t.Fatalf("expected packaging stage, got %q", stored.Stage)
	}
	if stored.Progress != 90 {
		t.Fatalf("expected progress 90, got %d", stored.Progress)
	}
	if stored.ErrorMessage != "save asset failed" {
		t.Fatalf("expected error message to be stored, got %q", stored.ErrorMessage)
	}
}

func TestWorkerRunRemovesDownloadedMediaOnSuccess(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	mediaPath, downloadDir := createDownloadedMedia(t)
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader(mediaPath),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if err := worker.Run(context.Background(), job.ID); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(downloadDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected download dir %q to be removed, got %v", downloadDir, err)
	}
}

func TestWorkerRunRemovesDownloadedMediaOnFailure(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	mediaPath, downloadDir := createDownloadedMedia(t)
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader(mediaPath),
		pipeline.FakeTranscriber([]pipeline.Segment{
			{
				Start:      "00:00:01.000",
				End:        "00:00:03.000",
				SourceText: "Hello",
			},
			{
				Start:      "00:00:03.000",
				End:        "00:00:05.000",
				SourceText: "World",
			},
		}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	err = worker.Run(context.Background(), job.ID)
	if err == nil {
		t.Fatal("expected Run to fail for misaligned translation output")
	}

	if _, err := os.Stat(downloadDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected download dir %q to be removed, got %v", downloadDir, err)
	}
}

func TestWorkerRunPendingLoopProcessesQueuedJobs(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- worker.RunPendingLoop(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		stored, getErr := svc.GetJob(context.Background(), job.ID)
		if getErr != nil {
			t.Fatalf("GetJob returned error: %v", getErr)
		}

		if stored.Status == jobs.StatusCompleted {
			cancel()
			break
		}

		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queued job %q to complete", job.ID)
		}

		time.Sleep(10 * time.Millisecond)
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation from RunPendingLoop, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for RunPendingLoop to exit")
	}

	asset, err := svc.GetAssetByVideoID(context.Background(), job.VideoID)
	if err != nil {
		t.Fatalf("GetAssetByVideoID returned error: %v", err)
	}
	if asset.TranslatedVTTPath == "" {
		t.Fatal("expected pending loop to persist translated subtitle asset")
	}
}

func TestWorkerRunPersistsAssetsPerJobForSameVideo(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "jobs.db"))
	if err != nil {
		t.Fatalf("db.Open returned error: %v", err)
	}
	defer sqlDB.Close()

	repo, err := jobs.NewSQLiteRepository(sqlDB)
	if err != nil {
		t.Fatalf("NewSQLiteRepository returned error: %v", err)
	}

	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	nowCalls := 0
	svc := jobs.NewService(repo, func() time.Time {
		nowCalls++
		return now.Add(time.Duration(nowCalls-1) * time.Nanosecond)
	})
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	firstJob, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("first CreateJob returned error: %v", err)
	}

	secondJob, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "ja-JP",
	})
	if err != nil {
		t.Fatalf("second CreateJob returned error: %v", err)
	}

	if err := worker.Run(context.Background(), firstJob.ID); err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	if err := worker.Run(context.Background(), secondJob.ID); err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	firstAsset, err := svc.GetAssetByJobID(context.Background(), firstJob.ID)
	if err != nil {
		t.Fatalf("GetAssetByJobID for first job returned error: %v", err)
	}
	secondAsset, err := svc.GetAssetByJobID(context.Background(), secondJob.ID)
	if err != nil {
		t.Fatalf("GetAssetByJobID for second job returned error: %v", err)
	}

	if firstAsset.TargetLanguage != "zh-CN" {
		t.Fatalf("expected first asset target language %q, got %q", "zh-CN", firstAsset.TargetLanguage)
	}
	if secondAsset.TargetLanguage != "ja-JP" {
		t.Fatalf("expected second asset target language %q, got %q", "ja-JP", secondAsset.TargetLanguage)
	}
	if firstAsset.TranslatedVTTPath == secondAsset.TranslatedVTTPath {
		t.Fatalf("expected distinct translated paths, got %q", firstAsset.TranslatedVTTPath)
	}
}

func createDownloadedMedia(t *testing.T) (string, string) {
	t.Helper()

	downloadDir, err := os.MkdirTemp("", "lets-sub-it-download-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}

	mediaPath := filepath.Join(downloadDir, "job-1.mp4")
	if err := os.WriteFile(mediaPath, []byte("media"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return mediaPath, downloadDir
}

func TestWorkerRunPendingLoopRequeuesInterruptedRunningJobs(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	worker := pipeline.NewWorker(
		svc,
		pipeline.FakeDownloader("/tmp/video.mp4"),
		pipeline.FakeTranscriber([]pipeline.Segment{{
			Start:      "00:00:01.000",
			End:        "00:00:03.000",
			SourceText: "Hello",
		}}),
		pipeline.FakeTranslator([]pipeline.Segment{{
			Start:          "00:00:01.000",
			End:            "00:00:03.000",
			SourceText:     "Hello",
			TranslatedText: "你好",
		}}),
		t.TempDir(),
	)

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if err := svc.UpdateProgress(context.Background(), job.ID, jobs.ProgressUpdate{
		Status:   jobs.StatusRunning,
		Stage:    jobs.StageDownloading,
		Progress: 10,
	}); err != nil {
		t.Fatalf("UpdateProgress returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- worker.RunPendingLoop(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		stored, getErr := svc.GetJob(context.Background(), job.ID)
		if getErr != nil {
			t.Fatalf("GetJob returned error: %v", getErr)
		}

		if stored.Status == jobs.StatusCompleted {
			cancel()
			break
		}

		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for interrupted running job %q to complete", job.ID)
		}

		time.Sleep(10 * time.Millisecond)
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation from RunPendingLoop, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for RunPendingLoop to exit")
	}
}

func TestClaimQueuedJobsClaimsEachJobOnce(t *testing.T) {
	repo := jobs.NewMemoryRepositoryForTest()
	svc := jobs.NewService(repo, func() time.Time {
		return time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	})

	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	claimedFirst, err := svc.ClaimQueuedJobs(context.Background())
	if err != nil {
		t.Fatalf("ClaimQueuedJobs returned error: %v", err)
	}
	if len(claimedFirst) != 1 {
		t.Fatalf("expected first claim to return 1 job, got %d", len(claimedFirst))
	}
	if claimedFirst[0].ID != job.ID {
		t.Fatalf("expected claimed job %q, got %q", job.ID, claimedFirst[0].ID)
	}

	claimedSecond, err := svc.ClaimQueuedJobs(context.Background())
	if err != nil {
		t.Fatalf("second ClaimQueuedJobs returned error: %v", err)
	}
	if len(claimedSecond) != 0 {
		t.Fatalf("expected second claim to return 0 jobs, got %d", len(claimedSecond))
	}
}

func TestSQLiteClaimQueuedJobsClaimsEachJobOnceAcrossConcurrentWorkers(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "jobs.db"))
	if err != nil {
		t.Fatalf("db.Open returned error: %v", err)
	}
	defer sqlDB.Close()

	repo, err := jobs.NewSQLiteRepository(sqlDB)
	if err != nil {
		t.Fatalf("NewSQLiteRepository returned error: %v", err)
	}

	now := time.Date(2026, 4, 20, 10, 0, 0, 123456789, time.UTC)
	svc := jobs.NewService(repo, func() time.Time { return now })
	job, err := svc.CreateJob(context.Background(), jobs.CreateJobInput{
		VideoID:        "abc123xyz00",
		YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
		TargetLanguage: "zh-CN",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	results := make(chan []jobs.Job, 2)
	errs := make(chan error, 2)
	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)

	claim := func() {
		defer done.Done()
		start.Wait()
		claimed, err := svc.ClaimQueuedJobs(context.Background())
		if err != nil {
			errs <- err
			return
		}
		results <- claimed
	}

	done.Add(2)
	go claim()
	go claim()
	start.Done()
	done.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("ClaimQueuedJobs returned error: %v", err)
		}
	}

	claimCount := 0
	for claimed := range results {
		for _, claimedJob := range claimed {
			if claimedJob.ID != job.ID {
				t.Fatalf("expected claimed job %q, got %q", job.ID, claimedJob.ID)
			}
			claimCount++
		}
	}
	if claimCount != 1 {
		t.Fatalf("expected job to be claimed once, got %d", claimCount)
	}

	stored, err := svc.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetJob returned error: %v", err)
	}
	if stored.Status != jobs.StatusRunning {
		t.Fatalf("expected claimed job status %q, got %q", jobs.StatusRunning, stored.Status)
	}
}
