package service

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric types.Metric) error
	GetMetric(_ context.Context, rowMetric types.Metric) (types.Metric, error)
	GetMetrics(_ context.Context) ([]types.Metric, error)
	UpdateMetrics(_ context.Context, metrics []types.Metric) error
	Ping(ctx context.Context) error
	Close() error
}

type Service struct {
	config *serverconfig.Config
	repo   Repository
	logger zerolog.Logger
}

func New(ctx context.Context, config *serverconfig.Config, repo Repository, logger zerolog.Logger) *Service {
	s := &Service{
		config: config,
		repo:   repo,
		logger: logger,
	}
	s.loadMetricsFromFile(ctx)
	return s
}
