package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lets-sub-it-api/internal/runner"
	"lets-sub-it-api/internal/store"
)

type syncRunner struct {
	runner *runner.MockRunner
}

func (r syncRunner) Start(ctx context.Context, job store.Job) error {
	return r.runner.Start(ctx, job)
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	testStore, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	handler := NewHandler(testStore, syncRunner{runner: runner.NewMockRunner(testStore)}, t.TempDir())
	return Routes(handler)
}

func TestPostJobsCreatesJobAndCompletesWithMockRunner(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	request := httptest.NewRequest(http.MethodPost, "/jobs", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		Job    jobResponse `json:"job"`
		Reused bool        `json:"reused"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Reused || payload.Job.VideoID != "abc123" || payload.Job.SourceLanguage != "ja" {
		t.Fatalf("payload = %+v", payload)
	}

	waitForJobCompleted(t, server, payload.Job.ID)
}

func TestPostJobsRejectsMissingSourceLanguage(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123","targetLanguage":"zh-CN"}`)
	request := httptest.NewRequest(http.MethodPost, "/jobs", body)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestSubtitleAssetReturnsAssetAfterCompletion(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://youtu.be/abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/jobs", body)
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	var createPayload struct {
		Job jobResponse `json:"job"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("Unmarshal create response error = %v", err)
	}

	waitForJobCompleted(t, server, createPayload.Job.ID)
	assetResponse := waitForAsset(t, server, createPayload.Job.ID)
	if !bytes.Contains(assetResponse.Body.Bytes(), []byte("/subtitle-files/"+createPayload.Job.ID+"/translated")) {
		t.Fatalf("asset body = %s", assetResponse.Body.String())
	}
}

func TestSubtitleFileServing(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://youtu.be/abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/jobs", body)
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", createResponse.Code, createResponse.Body.String())
	}

	var createPayload struct {
		Job jobResponse `json:"job"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("Unmarshal create response error = %v", err)
	}

	waitForJobCompleted(t, server, createPayload.Job.ID)

	fileRequest := httptest.NewRequest(http.MethodGet, "/subtitle-files/"+createPayload.Job.ID+"/translated", nil)
	fileResponse := httptest.NewRecorder()
	server.ServeHTTP(fileResponse, fileRequest)

	if fileResponse.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", fileResponse.Code, fileResponse.Body.String())
	}
	if got := fileResponse.Header().Get("Content-Type"); got != "text/vtt; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if !bytes.HasPrefix(fileResponse.Body.Bytes(), []byte("WEBVTT")) {
		t.Fatalf("body = %s", fileResponse.Body.String())
	}
}

func TestSubtitleFileRejectsInvalidMode(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/subtitle-files/missing-job/invalid", nil)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
}

func TestSubtitleFileRejectsPathOutsideJobDir(t *testing.T) {
	workingDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "translated.vtt")
	if err := os.WriteFile(outsidePath, []byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\noutside secret\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	handler := NewHandler(handlerWithAssetPath{
		job: store.Job{
			ID:         "job_1",
			WorkingDir: workingDir,
		},
		asset: store.SubtitleAsset{
			JobID:             "job_1",
			SourceVTTPath:     filepath.Join(workingDir, "source.vtt"),
			TranslatedVTTPath: outsidePath,
			BilingualVTTPath:  filepath.Join(workingDir, "bilingual.vtt"),
		},
	}, noopRunner{}, t.TempDir())
	server := Routes(handler)

	request := httptest.NewRequest(http.MethodGet, "/subtitle-files/job_1/translated", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte(outsidePath)) {
		t.Fatalf("body leaked outside path = %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("outside secret")) {
		t.Fatalf("body leaked secret content = %s", response.Body.String())
	}
}

func TestSubtitleFileRejectsSymlinkEscape(t *testing.T) {
	workingDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "translated.vtt")
	if err := os.WriteFile(outsidePath, []byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\noutside symlink secret\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	symlinkPath := filepath.Join(workingDir, "translated.vtt")
	if err := os.Symlink(outsidePath, symlinkPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	handler := NewHandler(handlerWithAssetPath{
		job: store.Job{
			ID:         "job_1",
			WorkingDir: workingDir,
		},
		asset: store.SubtitleAsset{
			JobID:             "job_1",
			SourceVTTPath:     filepath.Join(workingDir, "source.vtt"),
			TranslatedVTTPath: symlinkPath,
			BilingualVTTPath:  filepath.Join(workingDir, "bilingual.vtt"),
		},
	}, noopRunner{}, t.TempDir())
	server := Routes(handler)

	request := httptest.NewRequest(http.MethodGet, "/subtitle-files/job_1/translated", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte(outsidePath)) {
		t.Fatalf("body leaked outside path = %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("outside symlink secret")) {
		t.Fatalf("body leaked secret content = %s", response.Body.String())
	}
}

func TestSubtitleFileRejectsNonRegularFile(t *testing.T) {
	workingDir := t.TempDir()
	directoryPath := filepath.Join(workingDir, "translated.vtt")
	if err := os.Mkdir(directoryPath, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	handler := NewHandler(handlerWithAssetPath{
		job: store.Job{
			ID:         "job_1",
			WorkingDir: workingDir,
		},
		asset: store.SubtitleAsset{
			JobID:             "job_1",
			SourceVTTPath:     filepath.Join(workingDir, "source.vtt"),
			TranslatedVTTPath: directoryPath,
			BilingualVTTPath:  filepath.Join(workingDir, "bilingual.vtt"),
		},
	}, noopRunner{}, t.TempDir())
	server := Routes(handler)

	request := httptest.NewRequest(http.MethodGet, "/subtitle-files/job_1/translated", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte(directoryPath)) {
		t.Fatalf("body leaked directory path = %s", response.Body.String())
	}
}

func TestSubtitleFileRejectsSymlinkParentEscape(t *testing.T) {
	workingDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "translated.vtt")
	if err := os.WriteFile(outsidePath, []byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nparent escape secret\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	linkPath := filepath.Join(workingDir, "link")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	handler := NewHandler(handlerWithAssetPath{
		job: store.Job{
			ID:         "job_1",
			WorkingDir: workingDir,
		},
		asset: store.SubtitleAsset{
			JobID:             "job_1",
			SourceVTTPath:     filepath.Join(workingDir, "source.vtt"),
			TranslatedVTTPath: filepath.Join(linkPath, "translated.vtt"),
			BilingualVTTPath:  filepath.Join(workingDir, "bilingual.vtt"),
		},
	}, noopRunner{}, t.TempDir())
	server := Routes(handler)

	request := httptest.NewRequest(http.MethodGet, "/subtitle-files/job_1/translated", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte(outsidePath)) {
		t.Fatalf("body leaked outside path = %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("parent escape secret")) {
		t.Fatalf("body leaked secret content = %s", response.Body.String())
	}
}

func waitForJobCompleted(t *testing.T, server http.Handler, jobID string) {
	t.Helper()
	var last *httptest.ResponseRecorder
	for range 20 {
		request := httptest.NewRequest(http.MethodGet, "/jobs/"+jobID, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		last = response

		if response.Code != http.StatusOK {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		var payload struct {
			Job jobResponse `json:"job"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("Unmarshal job response error = %v", err)
		}
		if payload.Job.Status == store.StatusCompleted {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job never became completed, last response = %s", last.Body.String())
}

func waitForAsset(t *testing.T, server http.Handler, jobID string) *httptest.ResponseRecorder {
	t.Helper()
	var last *httptest.ResponseRecorder
	for range 20 {
		assetRequest := httptest.NewRequest(http.MethodGet, "/subtitle-assets?videoId=abc123&targetLanguage=zh-CN", nil)
		assetResponse := httptest.NewRecorder()
		server.ServeHTTP(assetResponse, assetRequest)
		last = assetResponse
		if assetResponse.Code == http.StatusOK && bytes.Contains(assetResponse.Body.Bytes(), []byte("/subtitle-files/"+jobID+"/translated")) {
			return assetResponse
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("asset never became available, last response = %s", last.Body.String())
	return nil
}

type handlerWithAssetPath struct {
	job   store.Job
	asset store.SubtitleAsset
}

func (h handlerWithAssetPath) CreateJob(job store.Job) error {
	return nil
}

func (h handlerWithAssetPath) FindJob(id string) (store.Job, error) {
	if id == h.job.ID {
		return h.job, nil
	}
	return store.Job{}, store.ErrNotFound
}

func (h handlerWithAssetPath) FindReusableJob(videoID string, targetLanguage string) (store.Job, error) {
	return store.Job{}, store.ErrNotFound
}

func (h handlerWithAssetPath) FindSubtitleAsset(videoID string, targetLanguage string) (store.SubtitleAsset, error) {
	return store.SubtitleAsset{}, store.ErrNotFound
}

func (h handlerWithAssetPath) FindSubtitleAssetByJobID(jobID string) (store.SubtitleAsset, error) {
	if jobID == h.asset.JobID {
		return h.asset, nil
	}
	return store.SubtitleAsset{}, store.ErrNotFound
}

type noopRunner struct{}

func (noopRunner) Start(ctx context.Context, job store.Job) error { return nil }
