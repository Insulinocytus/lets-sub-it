package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrJobNotFound = errors.New("job not found")
var ErrJobAlreadyExists = errors.New("job already exists")
var ErrAssetNotFound = errors.New("asset not found")

type Repository interface {
	InsertJob(ctx context.Context, job Job) error
	UpdateJob(ctx context.Context, job Job) error
	GetJob(ctx context.Context, id string) (Job, error)
	ListJobsByStatus(ctx context.Context, status Status) ([]Job, error)
	ClaimQueuedJobs(ctx context.Context) ([]Job, error)
	SaveAsset(ctx context.Context, asset SubtitleAsset) error
	GetAssetByVideoID(ctx context.Context, videoID string) (SubtitleAsset, error)
}

type inMemoryRepository struct {
	mu     sync.Mutex
	jobs   map[string]Job
	assets map[string]SubtitleAsset
}

func NewMemoryRepository() Repository {
	return newInMemoryRepository()
}

func newInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		jobs:   map[string]Job{},
		assets: map[string]SubtitleAsset{},
	}
}

func NewMemoryRepositoryForTest() Repository {
	return NewMemoryRepository()
}

func (r *inMemoryRepository) InsertJob(_ context.Context, job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jobs[job.ID]; ok {
		return ErrJobAlreadyExists
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *inMemoryRepository) UpdateJob(_ context.Context, job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jobs[job.ID]; !ok {
		return ErrJobNotFound
	}

	r.jobs[job.ID] = job
	return nil
}

func (r *inMemoryRepository) GetJob(_ context.Context, id string) (Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[id]
	if !ok {
		return Job{}, ErrJobNotFound
	}

	return job, nil
}

func (r *inMemoryRepository) ListJobsByStatus(_ context.Context, status Status) ([]Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		if job.Status == status {
			result = append(result, job)
		}
	}

	return result, nil
}

func (r *inMemoryRepository) ClaimQueuedJobs(_ context.Context) ([]Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var claimed []Job
	for id, job := range r.jobs {
		if job.Status != StatusQueued {
			continue
		}

		job.Status = StatusRunning
		job.Stage = StageQueued
		r.jobs[id] = job
		claimed = append(claimed, job)
	}

	return claimed, nil
}

func (r *inMemoryRepository) SaveAsset(_ context.Context, asset SubtitleAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.assets[asset.VideoID] = asset
	return nil
}

func (r *inMemoryRepository) GetAssetByVideoID(_ context.Context, videoID string) (SubtitleAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset, ok := r.assets[videoID]
	if !ok {
		return SubtitleAsset{}, ErrAssetNotFound
	}

	return asset, nil
}

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) (Repository, error) {
	repo := &sqliteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *sqliteRepository) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			video_id TEXT NOT NULL,
			youtube_url TEXT NOT NULL,
			target_language TEXT NOT NULL,
			status TEXT NOT NULL,
			stage TEXT NOT NULL,
			progress INTEGER NOT NULL,
			error_message TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS subtitle_assets (
			video_id TEXT PRIMARY KEY,
			job_id TEXT NOT NULL,
			source_vtt_path TEXT NOT NULL,
			translated_vtt_path TEXT NOT NULL,
			bilingual_vtt_path TEXT NOT NULL,
			source_language TEXT NOT NULL,
			target_language TEXT NOT NULL
		)`,
	}

	for _, statement := range statements {
		if _, err := r.db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func (r *sqliteRepository) InsertJob(ctx context.Context, job Job) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO jobs (
			id, video_id, youtube_url, target_language, status, stage, progress, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID,
		job.VideoID,
		job.YouTubeURL,
		job.TargetLanguage,
		string(job.Status),
		string(job.Stage),
		job.Progress,
		job.ErrorMessage,
		job.CreatedAt.UTC().Format(time.RFC3339Nano),
		job.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		if isSQLiteConstraintError(err) {
			return ErrJobAlreadyExists
		}

		return err
	}

	return nil
}

func (r *sqliteRepository) UpdateJob(ctx context.Context, job Job) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE jobs
		SET video_id = ?, youtube_url = ?, target_language = ?, status = ?, stage = ?, progress = ?, error_message = ?, created_at = ?, updated_at = ?
		WHERE id = ?`,
		job.VideoID,
		job.YouTubeURL,
		job.TargetLanguage,
		string(job.Status),
		string(job.Stage),
		job.Progress,
		job.ErrorMessage,
		job.CreatedAt.UTC().Format(time.RFC3339Nano),
		job.UpdatedAt.UTC().Format(time.RFC3339Nano),
		job.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

func (r *sqliteRepository) GetJob(ctx context.Context, id string) (Job, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			id, video_id, youtube_url, target_language, status, stage, progress, error_message, created_at, updated_at
		FROM jobs
		WHERE id = ?`,
		id,
	)

	job, err := scanJob(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, ErrJobNotFound
		}

		return Job{}, err
	}

	return job, nil
}

func (r *sqliteRepository) ListJobsByStatus(ctx context.Context, status Status) ([]Job, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			id, video_id, youtube_url, target_language, status, stage, progress, error_message, created_at, updated_at
		FROM jobs
		WHERE status = ?
		ORDER BY created_at ASC`,
		string(status),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}

	return result, rows.Err()
}

func (r *sqliteRepository) ClaimQueuedJobs(ctx context.Context) ([]Job, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	rows, err := conn.QueryContext(
		ctx,
		`SELECT
			id, video_id, youtube_url, target_language, status, stage, progress, error_message, created_at, updated_at
		FROM jobs
		WHERE status = ?
		ORDER BY created_at ASC`,
		string(StatusQueued),
	)
	if err != nil {
		return nil, err
	}

	var claimed []Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		job.Status = StatusRunning
		claimed = append(claimed, job)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, job := range claimed {
		if _, err := conn.ExecContext(
			ctx,
			`UPDATE jobs SET status = ? WHERE id = ?`,
			string(StatusRunning),
			job.ID,
		); err != nil {
			return nil, err
		}
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return nil, err
	}
	committed = true

	return claimed, nil
}

func (r *sqliteRepository) SaveAsset(ctx context.Context, asset SubtitleAsset) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO subtitle_assets (
			video_id, job_id, source_vtt_path, translated_vtt_path, bilingual_vtt_path, source_language, target_language
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(video_id) DO UPDATE SET
			job_id = excluded.job_id,
			source_vtt_path = excluded.source_vtt_path,
			translated_vtt_path = excluded.translated_vtt_path,
			bilingual_vtt_path = excluded.bilingual_vtt_path,
			source_language = excluded.source_language,
			target_language = excluded.target_language`,
		asset.VideoID,
		asset.JobID,
		asset.SourceVTTPath,
		asset.TranslatedVTTPath,
		asset.BilingualVTTPath,
		asset.SourceLanguage,
		asset.TargetLanguage,
	)

	return err
}

func (r *sqliteRepository) GetAssetByVideoID(ctx context.Context, videoID string) (SubtitleAsset, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			job_id, video_id, source_vtt_path, translated_vtt_path, bilingual_vtt_path, source_language, target_language
		FROM subtitle_assets
		WHERE video_id = ?`,
		videoID,
	)

	var asset SubtitleAsset
	err := row.Scan(
		&asset.JobID,
		&asset.VideoID,
		&asset.SourceVTTPath,
		&asset.TranslatedVTTPath,
		&asset.BilingualVTTPath,
		&asset.SourceLanguage,
		&asset.TargetLanguage,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SubtitleAsset{}, ErrAssetNotFound
		}

		return SubtitleAsset{}, err
	}

	return asset, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanJob(row scannable) (Job, error) {
	var job Job
	var status string
	var stage string
	var createdAt string
	var updatedAt string

	err := row.Scan(
		&job.ID,
		&job.VideoID,
		&job.YouTubeURL,
		&job.TargetLanguage,
		&status,
		&stage,
		&job.Progress,
		&job.ErrorMessage,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return Job{}, err
	}

	job.Status = Status(status)
	job.Stage = Stage(stage)

	job.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Job{}, fmt.Errorf("parse created_at: %w", err)
	}

	job.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Job{}, fmt.Errorf("parse updated_at: %w", err)
	}

	return job, nil
}

func isSQLiteConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
