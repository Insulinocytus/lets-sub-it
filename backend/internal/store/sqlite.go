package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(foreignKeyDSN(path)), &gorm.Config{
		Logger: newGormSlogLogger(200 * time.Millisecond),
	})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

type gormSlogLogger struct {
	slowThreshold time.Duration
	level         gormlogger.LogLevel
}

func newGormSlogLogger(slowThreshold time.Duration) gormlogger.Interface {
	return gormSlogLogger{
		slowThreshold: slowThreshold,
		level:         gormlogger.Info,
	}
}

func (l gormSlogLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.level = level
	return l
}

func (l gormSlogLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		slog.InfoContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l gormSlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		slog.WarnContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l gormSlogLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		slog.ErrorContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l gormSlogLogger) Trace(ctx context.Context, startedAt time.Time, fc func() (string, int64), err error) {
	if l.level == gormlogger.Silent {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	duration := time.Since(startedAt)
	_, rows := fc()
	attrs := []any{
		"duration_ms", duration.Milliseconds(),
		"rows", rows,
	}

	if err != nil && l.level >= gormlogger.Error {
		slog.ErrorContext(ctx, "database query failed", append(attrs, "error", err)...)
		return
	}
	if l.slowThreshold > 0 && duration > l.slowThreshold && l.level >= gormlogger.Warn {
		slog.WarnContext(ctx, "database query slow", append(attrs, "slow_threshold_ms", l.slowThreshold.Milliseconds())...)
		return
	}
	if l.level >= gormlogger.Info {
		slog.DebugContext(ctx, "database query", attrs...)
	}
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

func (s *Store) FindLatestJob(videoID string, targetLanguage string) (Job, error) {
	var job Job
	err := s.db.Where("video_id = ? AND target_language = ?", videoID, targetLanguage).
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

func (s *Store) FailInterruptedJobs(reason string) (int64, error) {
	updates := map[string]any{
		"status":        StatusFailed,
		"progress_text": "处理失败",
		"error_message": reason,
	}
	result := s.db.Model(&Job{}).
		Where("status NOT IN ?", []string{StatusCompleted, StatusFailed}).
		Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (s *Store) CreateSubtitleAsset(asset SubtitleAsset) error {
	return s.db.Create(&asset).Error
}

// DeleteSubtitleResult logically deletes every job for the video and target language and
// returns their working directories. Assets stay in place and are hidden through their job.
func (s *Store) DeleteSubtitleResult(videoID string, targetLanguage string) ([]string, error) {
	var jobs []Job
	if err := s.db.Where("video_id = ? AND target_language = ?", videoID, targetLanguage).Order("id").Find(&jobs).Error; err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(jobs))
	dirs := make([]string, len(jobs))
	for i, job := range jobs {
		ids[i] = job.ID
		dirs[i] = job.WorkingDir
	}
	if err := s.db.Where("id IN ?", ids).Delete(&Job{}).Error; err != nil {
		return nil, err
	}
	return dirs, nil
}

func (s *Store) FindSubtitleAsset(videoID string, targetLanguage string) (SubtitleAsset, error) {
	return s.findVisibleSubtitleAsset("subtitle_assets.video_id = ? AND subtitle_assets.target_language = ?", videoID, targetLanguage)
}

func (s *Store) FindSubtitleAssetByJobID(jobID string) (SubtitleAsset, error) {
	return s.findVisibleSubtitleAsset("subtitle_assets.job_id = ?", jobID)
}

// findVisibleSubtitleAsset only returns assets whose job is not deleted, so an asset written
// by a runner after its job was deleted never becomes visible.
func (s *Store) findVisibleSubtitleAsset(query string, args ...any) (SubtitleAsset, error) {
	var asset SubtitleAsset
	err := s.db.Joins("JOIN jobs ON jobs.id = subtitle_assets.job_id AND jobs.deleted_at IS NULL").
		Where(query, args...).
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubtitleAsset{}, ErrNotFound
		}
		return SubtitleAsset{}, err
	}
	return asset, nil
}
