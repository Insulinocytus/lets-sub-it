package http

import (
	"bytes"
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"lets-sub-it/backend/internal/jobs"
)

func TestCreateJobAndFetchJob(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/jobs",
		bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123xyz00","targetLanguage":"zh-CN"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, req)
	if createResp.Code != stdhttp.StatusAccepted {
		t.Fatalf("expected 202, got %d", createResp.Code)
	}

	var created JobResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("expected valid json body, got error: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected create response to contain job id")
	}
	if created.VideoID != "abc123xyz00" {
		t.Fatalf("expected video id %q, got %q", "abc123xyz00", created.VideoID)
	}
	if created.Status != string(jobs.StatusQueued) {
		t.Fatalf("expected status %q, got %q", jobs.StatusQueued, created.Status)
	}
	if created.Stage != string(jobs.StageQueued) {
		t.Fatalf("expected stage %q, got %q", jobs.StageQueued, created.Stage)
	}
	if created.Progress != 0 {
		t.Fatalf("expected progress 0, got %d", created.Progress)
	}

	getReq := httptest.NewRequest(stdhttp.MethodGet, "/api/jobs/"+created.ID, nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	if getResp.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.Code)
	}

	var fetched JobResponse
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("expected valid get job json body, got error: %v", err)
	}
	if fetched != created {
		t.Fatalf("expected fetched job %+v, got %+v", created, fetched)
	}
}

func TestCreateJobReturnsBadRequestForInvalidJSON(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/jobs",
		bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123xyz00","targetLanguage":"zh-CN"`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestCreateJobReturnsBadRequestForInvalidOrUnsupportedYouTubeURL(t *testing.T) {
	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "relative url",
			payload: `{"youtubeUrl":"/watch?v=abc123xyz00","targetLanguage":"zh-CN"}`,
		},
		{
			name:    "unsupported host",
			payload: `{"youtubeUrl":"https://example.com/watch?v=abc123xyz00","targetLanguage":"zh-CN"}`,
		},
		{
			name:    "invalid video id",
			payload: `{"youtubeUrl":"https://www.youtube.com/watch?v=short","targetLanguage":"zh-CN"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := newTestRouter()

			req := httptest.NewRequest(
				stdhttp.MethodPost,
				"/api/jobs",
				bytes.NewBufferString(tc.payload),
			)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != stdhttp.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.Code)
			}

			var body ErrorResponse
			if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid error response, got %v", err)
			}
			if !strings.Contains(strings.ToLower(body.Error), "youtube") {
				t.Fatalf("expected youtube validation error, got %q", body.Error)
			}
		})
	}
}

func TestCreateJobReturnsBadRequestForEmptyTargetLanguage(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/jobs",
		bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123xyz00","targetLanguage":"   "}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}

	var body ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid error response, got %v", err)
	}
	if !strings.Contains(body.Error, "target language is required") {
		t.Fatalf("expected target language validation error, got %q", body.Error)
	}
}

func TestGetJobReturnsNotFoundForMissingJobID(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/jobs/job-missing", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestGetAssetByVideoID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := jobs.NewService(jobs.NewMemoryRepositoryForTest(), testNow)
	if err := svc.SaveAsset(context.Background(), jobs.SubtitleAsset{
		JobID:             "job-1",
		VideoID:           "abc123xyz00",
		SourceVTTPath:     "abc123xyz00/job-1/source.vtt",
		TranslatedVTTPath: "abc123xyz00/job-1/translated.vtt",
		BilingualVTTPath:  "abc123xyz00/job-1/bilingual.vtt",
		SourceLanguage:    "en",
		TargetLanguage:    "zh-CN",
	}); err != nil {
		t.Fatalf("SaveAsset returned error: %v", err)
	}

	router := NewRouter(svc)
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/videos/abc123xyz00/subtitles", nil)
	req.Host = "localhost:8080"

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid asset json body, got error: %v", err)
	}

	if got := body["videoId"]; got != "abc123xyz00" {
		t.Fatalf("expected videoId %q, got %#v", "abc123xyz00", got)
	}

	subtitleURLs, ok := body["subtitleUrls"].(map[string]any)
	if !ok {
		t.Fatalf("expected subtitleUrls object, got %#v", body["subtitleUrls"])
	}

	if got := subtitleURLs["translated"]; got != "http://localhost:8080/assets/abc123xyz00/job-1/translated.vtt" {
		t.Fatalf("unexpected translated subtitle url: %#v", got)
	}
	if got := subtitleURLs["bilingual"]; got != "http://localhost:8080/assets/abc123xyz00/job-1/bilingual.vtt" {
		t.Fatalf("unexpected bilingual subtitle url: %#v", got)
	}
}

func TestGetAssetByJobID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := jobs.NewService(jobs.NewMemoryRepositoryForTest(), testNow)
	if err := svc.SaveAsset(context.Background(), jobs.SubtitleAsset{
		JobID:             "job-1",
		VideoID:           "abc123xyz00",
		SourceVTTPath:     "abc123xyz00/job-1/source.vtt",
		TranslatedVTTPath: "abc123xyz00/job-1/translated.vtt",
		BilingualVTTPath:  "abc123xyz00/job-1/bilingual.vtt",
		SourceLanguage:    "en",
		TargetLanguage:    "zh-CN",
	}); err != nil {
		t.Fatalf("SaveAsset returned error: %v", err)
	}

	router := NewRouter(svc)
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/jobs/job-1/subtitles", nil)
	req.Host = "localhost:8080"

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid asset json body, got error: %v", err)
	}

	if got := body["jobId"]; got != "job-1" {
		t.Fatalf("expected jobId %q, got %#v", "job-1", got)
	}

	subtitleURLs, ok := body["subtitleUrls"].(map[string]any)
	if !ok {
		t.Fatalf("expected subtitleUrls object, got %#v", body["subtitleUrls"])
	}

	if got := subtitleURLs["translated"]; got != "http://localhost:8080/assets/abc123xyz00/job-1/translated.vtt" {
		t.Fatalf("unexpected translated subtitle url: %#v", got)
	}
	if got := subtitleURLs["bilingual"]; got != "http://localhost:8080/assets/abc123xyz00/job-1/bilingual.vtt" {
		t.Fatalf("unexpected bilingual subtitle url: %#v", got)
	}
}

func TestStaticAssetsIncludeCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	assetsDir := t.TempDir()
	assetPath := assetsDir + "/abc123xyz00/job-1"
	if err := os.MkdirAll(assetPath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(assetPath+"/translated.vtt", []byte("WEBVTT\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	router := NewRouter(jobs.NewService(jobs.NewMemoryRepositoryForTest(), testNow))
	router.StaticFS("/assets", gin.Dir(assetsDir, false))

	req := httptest.NewRequest(stdhttp.MethodGet, "/assets/abc123xyz00/job-1/translated.vtt", nil)
	req.Header.Set("Origin", "https://www.youtube.com")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://www.youtube.com" {
		t.Fatalf("expected CORS origin header, got %q", got)
	}
}

func TestCreateJobAcceptsShortYouTubeURL(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/jobs",
		bytes.NewBufferString(`{"youtubeUrl":"https://youtu.be/abc123xyz00","targetLanguage":"zh-CN"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != stdhttp.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.Code)
	}

	var created JobResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("expected valid json body, got error: %v", err)
	}
	if created.VideoID != "abc123xyz00" {
		t.Fatalf("expected video id %q, got %q", "abc123xyz00", created.VideoID)
	}
	if created.YouTubeURL != "https://youtu.be/abc123xyz00" {
		t.Fatalf("expected youtube url to be preserved, got %q", created.YouTubeURL)
	}
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	svc := jobs.NewService(jobs.NewMemoryRepositoryForTest(), testNow)
	return NewRouter(svc)
}

func testNow() time.Time {
	return time.Unix(1713571200, 0)
}
