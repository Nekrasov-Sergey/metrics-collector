package service

import (
	"context"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric types.Metric) error
	GetMetric(_ context.Context, rowMetric types.Metric) (types.Metric, error)
	GetMetrics(_ context.Context) ([]types.Metric, error)
	UpdateMetrics(_ context.Context, metrics []types.Metric) error
}

type Service struct {
	repo   Repository
	config *serverconfig.Config
}

func New(ctx context.Context, repo Repository, config *serverconfig.Config) *Service {
	s := &Service{
		repo:   repo,
		config: config,
	}
	s.loadMetricsFromFile(ctx)
	return s
}
