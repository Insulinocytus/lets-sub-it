package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	writeError(w, http.StatusNotFound, "not_found", "subtitle file handler is added in Task 7")
}
