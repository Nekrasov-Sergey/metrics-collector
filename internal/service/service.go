package service

import (
	"context"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type RepoInterface interface {
	UpdateMetric(ctx context.Context, typ types.MetricType, name types.MetricName, value float64) error
	GetMetric(_ context.Context, typ types.MetricType, name types.MetricName) (metric types.Metric, err error)
	GetMetrics(_ context.Context) (metrics []types.Metric, err error)
}

type Service struct {
	repo RepoInterface
}

func New(repo RepoInterface) *Service {
	return &Service{repo: repo}
}
