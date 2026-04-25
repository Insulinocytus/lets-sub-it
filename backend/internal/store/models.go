package store

import "time"

const (
	StatusQueued       = "queued"
	StatusDownloading  = "downloading"
	StatusTranscribing = "transcribing"
	StatusTranslating  = "translating"
	StatusPackaging    = "packaging"
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
)

type Job struct {
	ID             string  `gorm:"primaryKey;column:id;not null"`
	VideoID        string  `gorm:"column:video_id;not null;index:idx_jobs_lookup"`
	YoutubeURL     string  `gorm:"column:youtube_url;not null"`
	SourceLanguage string  `gorm:"column:source_language;not null"`
	TargetLanguage string  `gorm:"column:target_language;not null;index:idx_jobs_lookup"`
	Status         string  `gorm:"column:status;not null;index:idx_jobs_lookup"`
	Stage          string  `gorm:"column:stage;not null"`
	ProgressText   string  `gorm:"column:progress_text;not null"`
	ErrorMessage   *string `gorm:"column:error_message"`
	Attempt        int     `gorm:"column:attempt;not null"`
	WorkingDir     string  `gorm:"column:working_dir;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time `gorm:"index:idx_jobs_lookup"`
}

type SubtitleAsset struct {
	JobID             string `gorm:"primaryKey;column:job_id;not null;index:idx_subtitle_assets_lookup"`
	Job               Job    `gorm:"foreignKey:JobID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	VideoID           string `gorm:"column:video_id;not null;index:idx_subtitle_assets_lookup"`
	TargetLanguage    string `gorm:"column:target_language;not null;index:idx_subtitle_assets_lookup"`
	SourceVTTPath     string `gorm:"column:source_vtt_path;not null"`
	TranslatedVTTPath string `gorm:"column:translated_vtt_path;not null"`
	BilingualVTTPath  string `gorm:"column:bilingual_vtt_path;not null"`
	SourceLanguage    string `gorm:"column:source_language;not null"`
	CreatedAt         time.Time
}

func NewJob(id string, videoID string, youtubeURL string, sourceLanguage string, targetLanguage string, workingDir string) Job {
	return Job{
		ID:             id,
		VideoID:        videoID,
		YoutubeURL:     youtubeURL,
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		Status:         StatusQueued,
		Stage:          StatusQueued,
		ProgressText:   "等待处理",
		Attempt:        1,
		WorkingDir:     workingDir,
	}
}
