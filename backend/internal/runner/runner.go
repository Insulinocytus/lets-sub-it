package runner

import (
	"context"

	"lets-sub-it-api/internal/store"
)

type Runner interface {
	Start(ctx context.Context, job store.Job) error
}

type Store interface {
	UpdateJobStatus(id string, status string, stage string, progressText string, errorMessage string) error
	CreateSubtitleAsset(asset store.SubtitleAsset) error
}
