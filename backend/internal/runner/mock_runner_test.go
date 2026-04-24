package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lets-sub-it-api/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	testStore, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("store.Migrate() error = %v", err)
	}
	return testStore
}

func TestMockRunnerCompletesJobAndWritesAssets(t *testing.T) {
	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	if err := NewMockRunner(testStore).Start(context.Background(), job); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	updatedJob, err := testStore.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if updatedJob.Status != store.StatusCompleted {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusCompleted)
	}

	asset, err := testStore.FindSubtitleAsset("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", err)
	}

	for _, filePath := range []string{asset.SourceVTTPath, asset.TranslatedVTTPath, asset.BilingualVTTPath} {
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatalf("os.ReadFile(%q) error = %v", filePath, readErr)
		}
		if !strings.HasPrefix(string(content), "WEBVTT") {
			t.Fatalf("%q content = %q, want WEBVTT prefix", filePath, string(content))
		}
	}

	if asset.JobID != "job_1" {
		t.Fatalf("JobID = %q", asset.JobID)
	}
}

func TestMockRunnerMarksCanceledJobAsFailed(t *testing.T) {
	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewMockRunner(testStore).Start(ctx, job)
	if err == nil {
		t.Fatal("Start() error = nil, want context canceled")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusDownloading {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusDownloading)
	}
	if updatedJob.ProgressText != "处理失败" {
		t.Fatalf("ProgressText = %q, want %q", updatedJob.ProgressText, "处理失败")
	}
	if updatedJob.ErrorMessage == nil {
		t.Fatal("ErrorMessage = nil, want non-nil")
	}
}

func TestMockRunnerFailsJobWhenCompletionUpdateFails(t *testing.T) {
	fakeStore := &recordingStore{
		failOnCompleted: errors.New("completed update failed"),
	}
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", t.TempDir())

	err := NewMockRunner(fakeStore).Start(context.Background(), job)
	if !errors.Is(err, fakeStore.failOnCompleted) {
		t.Fatalf("Start() error = %v, want %v", err, fakeStore.failOnCompleted)
	}

	if len(fakeStore.calls) == 0 {
		t.Fatal("expected UpdateJobStatus calls")
	}

	var sawCompletedUpdate bool
	var sawFailedUpdate bool
	for _, call := range fakeStore.calls {
		if call.status == store.StatusCompleted {
			sawCompletedUpdate = true
		}
		if call.status == store.StatusFailed {
			sawFailedUpdate = true
			if call.stage != store.StatusCompleted {
				t.Fatalf("failed stage = %q, want %q", call.stage, store.StatusCompleted)
			}
			if call.progressText != "处理失败" {
				t.Fatalf("failed progressText = %q, want %q", call.progressText, "处理失败")
			}
			if call.errorMessage != fakeStore.failOnCompleted.Error() {
				t.Fatalf("failed errorMessage = %q, want %q", call.errorMessage, fakeStore.failOnCompleted.Error())
			}
		}
	}

	if !sawCompletedUpdate {
		t.Fatal("expected completed status update attempt")
	}
	if !sawFailedUpdate {
		t.Fatal("expected failed status update attempt")
	}
	if len(fakeStore.assets) != 1 {
		t.Fatalf("CreateSubtitleAsset calls = %d, want 1", len(fakeStore.assets))
	}
}

type recordingStore struct {
	calls           []updateCall
	assets          []store.SubtitleAsset
	failOnCompleted error
}

type updateCall struct {
	id           string
	status       string
	stage        string
	progressText string
	errorMessage string
}

func (s *recordingStore) UpdateJobStatus(id string, status string, stage string, progressText string, errorMessage string) error {
	s.calls = append(s.calls, updateCall{
		id:           id,
		status:       status,
		stage:        stage,
		progressText: progressText,
		errorMessage: errorMessage,
	})
	if status == store.StatusCompleted && s.failOnCompleted != nil {
		return s.failOnCompleted
	}
	return nil
}

func (s *recordingStore) CreateSubtitleAsset(asset store.SubtitleAsset) error {
	s.assets = append(s.assets, asset)
	return nil
}
