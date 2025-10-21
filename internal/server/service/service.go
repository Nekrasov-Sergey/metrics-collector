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
	config     *serverconfig.Config
	memStorage Repository
}

func New(ctx context.Context, config *serverconfig.Config, memStorage Repository) *Service {
	s := &Service{
		config:     config,
		memStorage: memStorage,
	}
	s.loadMetricsFromFile(ctx)
	return s
}
