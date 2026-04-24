package api

import (
	"encoding/json"
	"net/http"
	"time"

	"lets-sub-it-api/internal/store"
)

type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type jobResponse struct {
	ID             string  `json:"id"`
	VideoID        string  `json:"videoId"`
	YoutubeURL     string  `json:"youtubeUrl"`
	SourceLanguage string  `json:"sourceLanguage"`
	TargetLanguage string  `json:"targetLanguage"`
	Status         string  `json:"status"`
	Stage          string  `json:"stage"`
	ProgressText   string  `json:"progressText"`
	ErrorMessage   *string `json:"errorMessage"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type assetResponse struct {
	JobID          string            `json:"jobId"`
	VideoID        string            `json:"videoId"`
	TargetLanguage string            `json:"targetLanguage"`
	SourceLanguage string            `json:"sourceLanguage"`
	Files          map[string]string `json:"files"`
	CreatedAt      string            `json:"createdAt"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorBody{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func toJobResponse(job store.Job) jobResponse {
	return jobResponse{
		ID:             job.ID,
		VideoID:        job.VideoID,
		YoutubeURL:     job.YoutubeURL,
		SourceLanguage: job.SourceLanguage,
		TargetLanguage: job.TargetLanguage,
		Status:         job.Status,
		Stage:          job.Stage,
		ProgressText:   job.ProgressText,
		ErrorMessage:   job.ErrorMessage,
		CreatedAt:      formatTime(job.CreatedAt),
		UpdatedAt:      formatTime(job.UpdatedAt),
	}
}

func toAssetResponse(asset store.SubtitleAsset) assetResponse {
	return assetResponse{
		JobID:          asset.JobID,
		VideoID:        asset.VideoID,
		TargetLanguage: asset.TargetLanguage,
		SourceLanguage: asset.SourceLanguage,
		Files: map[string]string{
			"source":     "/subtitle-files/" + asset.JobID + "/source",
			"translated": "/subtitle-files/" + asset.JobID + "/translated",
			"bilingual":  "/subtitle-files/" + asset.JobID + "/bilingual",
		},
		CreatedAt: formatTime(asset.CreatedAt),
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
