package service

import (
	"context"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric types.Metric) error
	GetMetric(_ context.Context, rowMetric types.Metric) (types.Metric, error)
	GetMetrics(_ context.Context) ([]types.Metric, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}
