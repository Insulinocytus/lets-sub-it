package runner

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"lets-sub-it-api/internal/store"
)

func TestRealRunnerCompletesJob(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "zh", "en", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	translator := fakeTranslator{translations: []string{"translated one"}}
	var calls []execCall
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, execCall{name: name, args: slices.Clone(args)})
		switch name {
		case "yt-dlp":
			return exec.CommandContext(ctx, "sh", "-c", "mkdir -p \"$1\" && printf fake-audio-data > \"$1/audio.mp3\"", "sh", jobDir)
		case "whisper-cli":
			outputPath := argValue(t, args, "--output")
			return exec.CommandContext(ctx, "sh", "-c", "printf 'WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nreal transcript\n' > \"$1\"", "sh", outputPath)
		default:
			return exec.CommandContext(ctx, "sh", "-c", "echo unexpected command >&2; exit 127")
		}
	}

	if err := NewRealRunner(testStore, 10*time.Minute, "tiny", "int8", translator).Start(context.Background(), job); err != nil {
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

	whisperCall := findExecCall(calls, "whisper-cli")
	if whisperCall == nil {
		t.Fatalf("exec calls = %#v, want whisper-cli call", calls)
	}
	sourcePath := filepath.Join(jobDir, "source.vtt")
	assertArg(t, whisperCall.args, "--input", audioPath)
	assertArg(t, whisperCall.args, "--output", sourcePath)
	assertArg(t, whisperCall.args, "--model", "tiny")
	assertArg(t, whisperCall.args, "--compute-type", "int8")
	assertArg(t, whisperCall.args, "--language", "zh")

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
	if strings.Contains(string(translatedContent), "mock 翻译") {
		t.Fatalf("translated.vtt content = %q, want no mock translation", string(translatedContent))
	}
	bilingualContent, readBilingualErr := os.ReadFile(asset.BilingualVTTPath)
	if readBilingualErr != nil {
		t.Fatalf("os.ReadFile(bilingual.vtt) error = %v", readBilingualErr)
	}
	if !strings.Contains(string(bilingualContent), "real transcript\ntranslated one") {
		t.Fatalf("bilingual.vtt content = %q, want transcript followed by translation", string(bilingualContent))
	}
}

func TestRealRunnerDownloadFailed(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=deleted", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Video unavailable' >&2 && exit 1")
	}

	err := NewRealRunner(testStore, 10*time.Minute, "small", "default", fakeTranslator{}).Start(context.Background(), job)
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

func TestRealRunnerMarksCanceledJobAsFailed(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "10")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRealRunner(testStore, 10*time.Minute, "small", "default", fakeTranslator{}).Start(ctx, job)
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
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		switch name {
		case "yt-dlp":
			return exec.CommandContext(ctx, "sh", "-c", "mkdir -p \"$1\" && printf fake-audio-data > \"$1/audio.mp3\"", "sh", jobDir)
		case "whisper-cli":
			return exec.CommandContext(ctx, "sh", "-c", "echo 'transcription failed: model download error' >&2 && exit 3")
		default:
			return exec.CommandContext(ctx, "sh", "-c", "echo unexpected command >&2; exit 127")
		}
	}

	err := NewRealRunner(testStore, 10*time.Minute, "small", "default", fakeTranslator{}).Start(context.Background(), job)
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
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		switch name {
		case "yt-dlp":
			return exec.CommandContext(ctx, "sh", "-c", "mkdir -p \"$1\" && printf fake-audio-data > \"$1/audio.mp3\"", "sh", jobDir)
		case "whisper-cli":
			outputPath := argValue(t, args, "--output")
			return exec.CommandContext(ctx, "sh", "-c", "printf 'WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nreal transcript\n' > \"$1\"", "sh", outputPath)
		default:
			return exec.CommandContext(ctx, "sh", "-c", "echo unexpected command >&2; exit 127")
		}
	}

	err := NewRealRunner(testStore, 10*time.Minute, "small", "default", fakeTranslator{err: errors.New("translation unavailable")}).Start(context.Background(), job)
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

type execCall struct {
	name string
	args []string
}

func findExecCall(calls []execCall, name string) *execCall {
	for i := range calls {
		if calls[i].name == name {
			return &calls[i]
		}
	}
	return nil
}

func assertArg(t *testing.T, args []string, flag string, want string) {
	t.Helper()
	if got := argValue(t, args, flag); got != want {
		t.Fatalf("%s arg = %q, want %q; args = %#v", flag, got, want, args)
	}
}

func argValue(t *testing.T, args []string, flag string) string {
	t.Helper()
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	t.Fatalf("missing %s arg in %#v", flag, args)
	return ""
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
