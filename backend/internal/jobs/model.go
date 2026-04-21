package jobs

import "time"

type Status string

type Stage string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

const (
	StageQueued       Stage = "queued"
	StageDownloading  Stage = "downloading"
	StageTranscribing Stage = "transcribing"
	StageTranslating  Stage = "translating"
	StagePackaging    Stage = "packaging"
	StageCompleted    Stage = "completed"
)

type Job struct {
	ID             string
	VideoID        string
	YouTubeURL     string
	TargetLanguage string
	Status         Status
	Stage          Stage
	Progress       int
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SubtitleAsset struct {
	JobID             string
	VideoID           string
	SourceVTTPath     string
	TranslatedVTTPath string
	BilingualVTTPath  string
	SourceLanguage    string
	TargetLanguage    string
}

type CreateJobInput struct {
	VideoID        string
	YouTubeURL     string
	TargetLanguage string
}

type ProgressUpdate struct {
	Status   Status
	Stage    Stage
	Progress int
	Error    string
}
