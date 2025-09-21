package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
)

func (s *Service) UpdateMetric(ctx context.Context, typ types.MetricType, name types.MetricName, value float64) error {
	if typ == types.Counter {
		metric, err := s.GetMetric(ctx, typ, name)
		if err != nil && !errors.Is(err, errcodes.ErrMetricNotFound) {
			return err
		}
		value += metric.Value
	}
	return s.repo.UpdateMetric(ctx, typ, name, value)
}

func (s *Service) GetMetric(ctx context.Context, typ types.MetricType, name types.MetricName) (metric types.Metric, err error) {
	return s.repo.GetMetric(ctx, typ, name)
}

func (s *Service) GetMetrics(ctx context.Context) (metrics []types.Metric, err error) {
	return s.repo.GetMetrics(ctx)
}
