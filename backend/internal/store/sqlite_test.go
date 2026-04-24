package store

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return store
}

func TestStoreCreatesAndFindsJob(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")

	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	found, err := store.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if found.VideoID != "abc123" || found.SourceLanguage != "ja" || found.TargetLanguage != "zh-CN" {
		t.Fatalf("found job = %+v", found)
	}
}

func TestStoreFindsReusableJob(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	found, err := store.FindReusableJob("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindReusableJob() error = %v", err)
	}
	if found.ID != "job_1" {
		t.Fatalf("reused job id = %q", found.ID)
	}

	if err := store.UpdateJobStatus("job_1", StatusFailed, StatusTranscribing, "失败", "boom"); err != nil {
		t.Fatalf("UpdateJobStatus() error = %v", err)
	}
	if _, err := store.FindReusableJob("abc123", "zh-CN"); err == nil {
		t.Fatal("FindReusableJob() expected not found for failed job")
	}
}

func TestStoreCreatesAndFindsSubtitleAsset(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	asset := SubtitleAsset{
		JobID:             "job_1",
		VideoID:           "abc123",
		TargetLanguage:    "zh-CN",
		SourceLanguage:    "ja",
		SourceVTTPath:     "/tmp/job_1/source.vtt",
		TranslatedVTTPath: "/tmp/job_1/translated.vtt",
		BilingualVTTPath:  "/tmp/job_1/bilingual.vtt",
	}

	if err := store.CreateSubtitleAsset(asset); err != nil {
		t.Fatalf("CreateSubtitleAsset() error = %v", err)
	}
	found, err := store.FindSubtitleAsset("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", err)
	}
	if found.JobID != "job_1" || found.SourceLanguage != "ja" {
		t.Fatalf("asset = %+v", found)
	}
}
