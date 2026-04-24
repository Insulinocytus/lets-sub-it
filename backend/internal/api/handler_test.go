package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	assetResponse := waitForAsset(t, server, createPayload.Job.ID)
	if !bytes.Contains(assetResponse.Body.Bytes(), []byte("/subtitle-files/"+createPayload.Job.ID+"/translated")) {
		t.Fatalf("asset body = %s", assetResponse.Body.String())
	}
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
