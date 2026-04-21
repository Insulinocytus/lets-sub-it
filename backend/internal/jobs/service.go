package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidCreateJobInput = errors.New("invalid create job input")
	ErrInvalidProgressUpdate = errors.New("invalid progress update")
)

type Service struct {
	repo   Repository
	nowFn  func() time.Time
	nextID func(time.Time) string
}

func NewService(repo Repository, nowFn func() time.Time) *Service {
	if nowFn == nil {
		nowFn = time.Now
	}

	return &Service{
		repo:  repo,
		nowFn: nowFn,
		nextID: func(now time.Time) string {
			return fmt.Sprintf("job-%d", now.UnixNano())
		},
	}
}

func (s *Service) CreateJob(ctx context.Context, input CreateJobInput) (Job, error) {
	input.VideoID = strings.TrimSpace(input.VideoID)
	input.YouTubeURL = strings.TrimSpace(input.YouTubeURL)
	input.TargetLanguage = strings.TrimSpace(input.TargetLanguage)

	if input.VideoID == "" {
		return Job{}, fmt.Errorf("%w: video id is required", ErrInvalidCreateJobInput)
	}
	if input.YouTubeURL == "" {
		return Job{}, fmt.Errorf("%w: youtube url is required", ErrInvalidCreateJobInput)
	}
	if input.TargetLanguage == "" {
		return Job{}, fmt.Errorf("%w: target language is required", ErrInvalidCreateJobInput)
	}

	now := s.nowFn()
	job := Job{
		ID:             s.nextID(now),
		VideoID:        input.VideoID,
		YouTubeURL:     input.YouTubeURL,
		TargetLanguage: input.TargetLanguage,
		Status:         StatusQueued,
		Stage:          StageQueued,
		Progress:       0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.InsertJob(ctx, job); err != nil {
		return Job{}, err
	}

	return job, nil
}

func (s *Service) UpdateProgress(ctx context.Context, jobID string, update ProgressUpdate) error {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return fmt.Errorf("%w: job id is required", ErrInvalidProgressUpdate)
	}
	if update.Status == "" {
		return fmt.Errorf("%w: status is required", ErrInvalidProgressUpdate)
	}
	if update.Stage == "" {
		return fmt.Errorf("%w: stage is required", ErrInvalidProgressUpdate)
	}
	if update.Progress < 0 || update.Progress > 100 {
		return fmt.Errorf("%w: progress must be between 0 and 100", ErrInvalidProgressUpdate)
	}

	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Status = update.Status
	job.Stage = update.Stage
	job.Progress = update.Progress
	job.ErrorMessage = update.Error
	job.UpdatedAt = s.nowFn()

	return s.repo.UpdateJob(ctx, job)
}

func (s *Service) GetJob(ctx context.Context, id string) (Job, error) {
	return s.repo.GetJob(ctx, id)
}

func (s *Service) ListJobsByStatus(ctx context.Context, status Status) ([]Job, error) {
	return s.repo.ListJobsByStatus(ctx, status)
}

func (s *Service) ClaimQueuedJobs(ctx context.Context) ([]Job, error) {
	return s.repo.ClaimQueuedJobs(ctx)
}

func (s *Service) SaveAsset(ctx context.Context, asset SubtitleAsset) error {
	return s.repo.SaveAsset(ctx, asset)
}

func (s *Service) GetAssetByVideoID(ctx context.Context, videoID string) (SubtitleAsset, error) {
	return s.repo.GetAssetByVideoID(ctx, strings.TrimSpace(videoID))
}
