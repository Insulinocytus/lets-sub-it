package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"lets-sub-it-api/internal/store"
)

type Store interface {
	CreateJob(job store.Job) error
	FindJob(id string) (store.Job, error)
	FindReusableJob(videoID string, targetLanguage string) (store.Job, error)
	FindSubtitleAsset(videoID string, targetLanguage string) (store.SubtitleAsset, error)
	FindSubtitleAssetByJobID(jobID string) (store.SubtitleAsset, error)
}

type Runner interface {
	Start(ctx context.Context, job store.Job) error
}

type Handler struct {
	store   Store
	runner  Runner
	workDir string
}

func NewHandler(store Store, runner Runner, workDir string) *Handler {
	return &Handler{
		store:   store,
		runner:  runner,
		workDir: workDir,
	}
}

type createJobRequest struct {
	YoutubeURL     string `json:"youtubeUrl"`
	SourceLanguage string `json:"sourceLanguage"`
	TargetLanguage string `json:"targetLanguage"`
}

func (h *Handler) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}

	var request createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	if request.YoutubeURL == "" || request.SourceLanguage == "" || request.TargetLanguage == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "youtubeUrl, sourceLanguage, and targetLanguage are required")
		return
	}

	videoID, err := ParseVideoID(request.YoutubeURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_youtube_url", "unsupported YouTube URL")
		return
	}

	reusableJob, err := h.store.FindReusableJob(videoID, request.TargetLanguage)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"job":    toJobResponse(reusableJob),
			"reused": true,
		})
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query reusable job")
		return
	}

	jobID := "job_" + strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")
	job := store.NewJob(
		jobID,
		videoID,
		request.YoutubeURL,
		request.SourceLanguage,
		request.TargetLanguage,
		filepath.Join(h.workDir, jobID),
	)

	if err := h.store.CreateJob(job); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create job")
		return
	}

	createdJob, err := h.store.FindJob(job.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query created job")
		return
	}

	go func() {
		_ = h.runner.Start(context.Background(), createdJob)
	}()

	writeJSON(w, http.StatusCreated, map[string]any{
		"job":    toJobResponse(createdJob),
		"reused": false,
	})
}

func (h *Handler) handleJobByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if jobID == "" || strings.Contains(jobID, "/") {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}

	job, err := h.store.FindJob(jobID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query job")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"job": toJobResponse(job),
	})
}

func (h *Handler) handleSubtitleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}

	videoID := r.URL.Query().Get("videoId")
	targetLanguage := r.URL.Query().Get("targetLanguage")
	if videoID == "" || targetLanguage == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "videoId and targetLanguage are required")
		return
	}

	asset, err := h.store.FindSubtitleAsset(videoID, targetLanguage)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{
			"asset": nil,
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query subtitle asset")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"asset": toAssetResponse(asset),
	})
}

func (h *Handler) handleSubtitleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/subtitle-files/")
	jobID, mode, ok := strings.Cut(path, "/")
	if !ok || jobID == "" || strings.Contains(jobID, "/") || strings.Contains(mode, "/") {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}

	expectedFileName, ok := subtitleFileNameForMode(mode)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_mode", "mode must be source, translated, or bilingual")
		return
	}

	asset, err := h.store.FindSubtitleAssetByJobID(jobID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query subtitle asset")
		return
	}

	job, err := h.store.FindJob(jobID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query job")
		return
	}

	var filePath string
	switch mode {
	case "source":
		filePath = asset.SourceVTTPath
	case "translated":
		filePath = asset.TranslatedVTTPath
	case "bilingual":
		filePath = asset.BilingualVTTPath
	}

	if !subtitleFilePathAllowed(job.WorkingDir, filePath, expectedFileName) {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}

	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	info, err := os.Lstat(filePath)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	if info.Mode()&os.ModeSymlink != 0 {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}

	info, err = os.Stat(filePath)
	if err != nil || !info.Mode().IsRegular() {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	http.ServeFile(w, r, filePath)
}

func subtitleFileNameForMode(mode string) (string, bool) {
	switch mode {
	case "source":
		return "source.vtt", true
	case "translated":
		return "translated.vtt", true
	case "bilingual":
		return "bilingual.vtt", true
	default:
		return "", false
	}
}

func subtitleFilePathAllowed(workingDir string, selectedPath string, expectedFileName string) bool {
	cleanWorkingDir := filepath.Clean(workingDir)
	cleanSelectedPath := filepath.Clean(selectedPath)
	if filepath.Base(cleanSelectedPath) != expectedFileName {
		return false
	}

	rel, err := filepath.Rel(cleanWorkingDir, cleanSelectedPath)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
