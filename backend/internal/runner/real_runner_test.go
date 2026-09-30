package runner

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lets-sub-it-api/internal/store"
)

func TestRealRunnerCompletesJob(t *testing.T) {
	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "zh", "en", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	translator := fakeTranslator{translations: []string{"translated one"}}
	transcriber := &fakeTranscriber{}
	useFakeYtDlp(t, ytDlpWritesAudio)

	if err := NewRealRunner(testStore, 10*time.Minute, "tiny", transcriber, translator).Start(context.Background(), job); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	updatedJob, err := testStore.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if updatedJob.Status != store.StatusCompleted {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusCompleted)
	}

	audioPath := filepath.Join(jobDir, "audio.mp3")
	data, readErr := os.ReadFile(audioPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(audio.mp3) error = %v", readErr)
	}
	if len(data) == 0 {
		t.Fatal("audio.mp3 is empty")
	}

	asset, assetErr := testStore.FindSubtitleAsset("abc123", "en")
	if assetErr != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", assetErr)
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
	sourceContent, readSourceErr := os.ReadFile(asset.SourceVTTPath)
	if readSourceErr != nil {
		t.Fatalf("os.ReadFile(source.vtt) error = %v", readSourceErr)
	}
	if !strings.Contains(string(sourceContent), "real transcript") {
		t.Fatalf("source.vtt content = %q, want real transcript", string(sourceContent))
	}
	translatedContent, readTranslatedErr := os.ReadFile(asset.TranslatedVTTPath)
	if readTranslatedErr != nil {
		t.Fatalf("os.ReadFile(translated.vtt) error = %v", readTranslatedErr)
	}
	if !strings.Contains(string(translatedContent), "translated one") {
		t.Fatalf("translated.vtt content = %q, want translated one", string(translatedContent))
	}
	bilingualContent, readBilingualErr := os.ReadFile(asset.BilingualVTTPath)
	if readBilingualErr != nil {
		t.Fatalf("os.ReadFile(bilingual.vtt) error = %v", readBilingualErr)
	}
	if !strings.Contains(string(bilingualContent), "real transcript\ntranslated one") {
		t.Fatalf("bilingual.vtt content = %q, want transcript followed by translation", string(bilingualContent))
	}
}

func TestRealRunnerLogsJobLifecycle(t *testing.T) {
	var output bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(originalLogger) })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, ytDlpWritesAudio)

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, fakeTranslator{translations: []string{"translated"}}).Start(context.Background(), job)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	logs := output.String()
	for _, want := range []string{
		`"msg":"job started"`,
		`"msg":"job stage started","job_id":"job_1","video_id":"abc123","stage":"downloading"`,
		`"msg":"job stage completed","job_id":"job_1","video_id":"abc123","stage":"downloading"`,
		`"msg":"job stage started","job_id":"job_1","video_id":"abc123","stage":"transcribing"`,
		`"msg":"job stage started","job_id":"job_1","video_id":"abc123","stage":"translating"`,
		`"msg":"job stage started","job_id":"job_1","video_id":"abc123","stage":"packaging"`,
		`"msg":"job completed","job_id":"job_1","video_id":"abc123"`,
	} {
		if !strings.Contains(logs, want) {
			t.Fatalf("logs = %s\nwant containing %s", logs, want)
		}
	}
}

func TestRealRunnerDownloadFailed(t *testing.T) {
	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=deleted", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, "fail:ERROR: Video unavailable")

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, fakeTranslator{}).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want error")
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
	if updatedJob.ErrorMessage == nil || !strings.Contains(*updatedJob.ErrorMessage, "Video unavailable") {
		t.Fatalf("ErrorMessage = %v, want containing 'Video unavailable'", updatedJob.ErrorMessage)
	}
}

func TestRealRunnerTranslationFailureDoesNotLogProviderBody(t *testing.T) {
	var output bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(originalLogger) })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, ytDlpWritesAudio)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "secret-key upstream echoed prompt: cue text", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	translator := NewChatTranslator(server.URL, "secret-key", "test-model", time.Second, server.Client())

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, translator).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want translation error")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}

	logs := output.String()
	errorMessage := ""
	if updatedJob.ErrorMessage != nil {
		errorMessage = *updatedJob.ErrorMessage
	}
	for _, leaked := range []string{"secret-key", "upstream echoed prompt", "cue text"} {
		if strings.Contains(logs, leaked) {
			t.Fatalf("logs leaked %q: %s", leaked, logs)
		}
		if strings.Contains(errorMessage, leaked) {
			t.Fatalf("job error message leaked %q: %s", leaked, errorMessage)
		}
	}
}

func TestRealRunnerMarksCanceledJobAsFailed(t *testing.T) {
	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, ytDlpHangs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, fakeTranslator{}).Start(ctx, job)
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
}

func TestRealRunnerTranscriptionFailed(t *testing.T) {
	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, ytDlpWritesAudio)

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{err: errors.New("model download error")}, fakeTranslator{}).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want transcription error")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusTranscribing {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusTranscribing)
	}
	if updatedJob.ErrorMessage == nil || !strings.Contains(*updatedJob.ErrorMessage, "model download error") {
		t.Fatalf("ErrorMessage = %v, want containing model download error", updatedJob.ErrorMessage)
	}
}

func TestRealRunnerTranslationFailed(t *testing.T) {
	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	useFakeYtDlp(t, ytDlpWritesAudio)

	err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, fakeTranslator{err: errors.New("translation unavailable")}).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want translation error")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusTranslating {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusTranslating)
	}
	if updatedJob.ErrorMessage == nil || !strings.Contains(*updatedJob.ErrorMessage, "translation unavailable") {
		t.Fatalf("ErrorMessage = %v, want containing translation unavailable", updatedJob.ErrorMessage)
	}
}

func TestRealRunnerDiscardsOutputOfJobDeletedWhileRunning(t *testing.T) {
	testStore := openTestStore(t)
	jobDir := filepath.Join(t.TempDir(), "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)
	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	useFakeYtDlp(t, ytDlpWritesAudio)
	translator := deletingTranslator{delete: func() {
		if _, err := testStore.DeleteSubtitleResult("abc123", "zh"); err != nil {
			t.Fatalf("DeleteSubtitleResult() error = %v", err)
		}
	}}

	if err := NewRealRunner(testStore, 10*time.Minute, "small", &fakeTranscriber{}, translator).Start(context.Background(), job); err != nil {
		t.Fatalf("Start() error = %v, want nil for a deleted job", err)
	}

	if _, err := os.Stat(jobDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("job dir stat error = %v, want removed", err)
	}
	if _, err := testStore.FindSubtitleAsset("abc123", "zh"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("FindSubtitleAsset() error = %v, want ErrNotFound", err)
	}
}

// deletingTranslator deletes the job's Subtitle Result mid-run, then translates normally.
type deletingTranslator struct {
	delete func()
}

func (t deletingTranslator) Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error) {
	t.delete()
	return fakeTranslator{}.Translate(ctx, cues, sourceLanguage, targetLanguage)
}

type fakeTranscriber struct {
	err error
}

func (t *fakeTranscriber) Transcribe(ctx context.Context, request TranscriptionRequest) error {
	if t.err != nil {
		return t.err
	}
	if err := os.MkdirAll(filepath.Dir(request.SourcePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(request.SourcePath, []byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nreal transcript\n"), 0o644)
}

type fakeTranslator struct {
	translations []string
	err          error
}

func (t fakeTranslator) Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error) {
	if t.err != nil {
		return nil, t.err
	}
	if len(t.translations) > 0 {
		return t.translations, nil
	}
	translations := make([]string, len(cues))
	for i := range cues {
		translations[i] = "translated"
	}
	return translations, nil
}
