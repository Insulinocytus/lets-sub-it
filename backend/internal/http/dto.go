package http

import "lets-sub-it/backend/internal/jobs"

type CreateJobRequest struct {
	YouTubeURL     string `json:"youtubeUrl"`
	TargetLanguage string `json:"targetLanguage"`
}

type JobResponse struct {
	ID             string `json:"id"`
	VideoID        string `json:"videoId"`
	YouTubeURL     string `json:"youtubeUrl"`
	TargetLanguage string `json:"targetLanguage"`
	Status         string `json:"status"`
	Stage          string `json:"stage"`
	Progress       int    `json:"progress"`
	ErrorMessage   string `json:"errorMessage"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SubtitleAssetResponse struct {
	JobID          string       `json:"jobId"`
	VideoID        string       `json:"videoId"`
	SourceLanguage string       `json:"sourceLanguage"`
	TargetLanguage string       `json:"targetLanguage"`
	SubtitleURLs   SubtitleURLs `json:"subtitleUrls"`
}

type SubtitleURLs struct {
	Source     string `json:"source"`
	Translated string `json:"translated"`
	Bilingual  string `json:"bilingual"`
}

func toJobResponse(job jobs.Job) JobResponse {
	return JobResponse{
		ID:             job.ID,
		VideoID:        job.VideoID,
		YouTubeURL:     job.YouTubeURL,
		TargetLanguage: job.TargetLanguage,
		Status:         string(job.Status),
		Stage:          string(job.Stage),
		Progress:       job.Progress,
		ErrorMessage:   job.ErrorMessage,
	}
}

func toSubtitleAssetResponse(asset jobs.SubtitleAsset, sourceURL string, translatedURL string, bilingualURL string) SubtitleAssetResponse {
	return SubtitleAssetResponse{
		JobID:          asset.JobID,
		VideoID:        asset.VideoID,
		SourceLanguage: asset.SourceLanguage,
		TargetLanguage: asset.TargetLanguage,
		SubtitleURLs: SubtitleURLs{
			Source:     sourceURL,
			Translated: translatedURL,
			Bilingual:  bilingualURL,
		},
	}
}
