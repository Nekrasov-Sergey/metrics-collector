package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (s *Service) UpdateMetric(ctx context.Context, metric types.Metric) error {
	if metric.MType == types.Counter {
		counterMetric, err := s.GetMetric(ctx, metric)
		if err != nil && !errors.Is(err, errcodes.ErrMetricNotFound) {
			return err
		}
		*metric.Delta += utils.Deref(counterMetric.Delta)
	}
	return s.repo.UpdateMetric(ctx, metric)
}

func (s *Service) GetMetric(ctx context.Context, rowMetric types.Metric) (metric types.Metric, err error) {
	return s.repo.GetMetric(ctx, rowMetric)
}

func (s *Service) GetMetrics(ctx context.Context) (metrics []types.Metric, err error) {
	return s.repo.GetMetrics(ctx)
}
