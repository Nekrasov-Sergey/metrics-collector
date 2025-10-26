package memstorage

import (
	"context"
	"sort"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (m *MemStorage) UpdateMetric(ctx context.Context, metric types.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[metric.Name] = metric
	switch metric.MType {
	case types.Gauge:
		logger.C(ctx).Info().
			Str("тип", string(types.Gauge)).
			Str("имя", string(metric.Name)).
			Float64("значение", utils.Deref(metric.Value)).
			Msg("Обновлённая метрика")
	case types.Counter:
		logger.C(ctx).Info().
			Str("тип", string(types.Counter)).
			Str("имя", string(metric.Name)).
			Int64("значение", utils.Deref(metric.Delta)).
			Msg("Обновлённая метрика")
	}

	return nil
}

func (m *MemStorage) GetMetric(_ context.Context, rowMetric types.Metric) (types.Metric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric, ok := m.metrics[rowMetric.Name]
	if !ok {
		return types.Metric{}, errcodes.ErrMetricNotFound
	}

	if metric.MType != rowMetric.MType {
		return types.Metric{}, errcodes.ErrMetricNotFound
	}

	return metric, nil
}

func (m *MemStorage) GetMetrics(_ context.Context) ([]types.Metric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make([]types.Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}

	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics, nil
}

func (m *MemStorage) UpdateMetrics(_ context.Context, metrics []types.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range metrics {
		if metric.MType == types.Counter {
			if counterMetric, ok := m.metrics[metric.Name]; ok {
				*metric.Delta += utils.Deref(counterMetric.Delta)
			}
		}
		m.metrics[metric.Name] = metric
	}

	return nil
}
