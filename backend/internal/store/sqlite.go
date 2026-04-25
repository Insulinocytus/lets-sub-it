package store

import (
	"errors"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(foreignKeyDSN(path)), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func foreignKeyDSN(path string) string {
	if strings.Contains(path, "?") {
		return path + "&_foreign_keys=on"
	}
	return path + "?_foreign_keys=on"
}

func (s *Store) CreateJob(job Job) error {
	return s.db.Create(&job).Error
}

func (s *Store) FindJob(id string) (Job, error) {
	var job Job
	if err := s.db.First(&job, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Job{}, ErrNotFound
		}
		return Job{}, err
	}
	return job, nil
}

func (s *Store) FindReusableJob(videoID string, targetLanguage string) (Job, error) {
	var job Job
	err := s.db.Where("video_id = ? AND target_language = ? AND status <> ?", videoID, targetLanguage, StatusFailed).
		Order("updated_at DESC").
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Job{}, ErrNotFound
		}
		return Job{}, err
	}
	return job, nil
}

func (s *Store) UpdateJobStatus(id string, status string, stage string, progressText string, errorMessage string) error {
	updates := map[string]any{
		"status":        status,
		"stage":         stage,
		"progress_text": progressText,
	}
	if errorMessage == "" {
		updates["error_message"] = nil
	} else {
		updates["error_message"] = errorMessage
	}

	result := s.db.Model(&Job{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateSubtitleAsset(asset SubtitleAsset) error {
	return s.db.Create(&asset).Error
}

func (s *Store) FindSubtitleAsset(videoID string, targetLanguage string) (SubtitleAsset, error) {
	var asset SubtitleAsset
	if err := s.db.First(&asset, "video_id = ? AND target_language = ?", videoID, targetLanguage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubtitleAsset{}, ErrNotFound
		}
		return SubtitleAsset{}, err
	}
	return asset, nil
}

func (s *Store) FindSubtitleAssetByJobID(jobID string) (SubtitleAsset, error) {
	var asset SubtitleAsset
	if err := s.db.First(&asset, "job_id = ?", jobID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubtitleAsset{}, ErrNotFound
		}
		return SubtitleAsset{}, err
	}
	return asset, nil
}
