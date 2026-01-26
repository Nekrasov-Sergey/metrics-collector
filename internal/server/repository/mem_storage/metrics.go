package memstorage

import (
	"context"
	"sort"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (m *MemStorage) UpdateMetric(ctx context.Context, metric *types.Metric) error {
	if metric.MType == types.Counter {
		counterMetric, err := m.GetMetric(ctx, metric)
		if err != nil && !errors.Is(err, errcodes.ErrMetricNotFound) {
			return err
		}
		if counterMetric != nil {
			*metric.Delta += utils.Deref(counterMetric.Delta)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[metric.Name] = metric
	return nil
}

func (m *MemStorage) GetMetric(_ context.Context, rawMetric *types.Metric) (*types.Metric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric, ok := m.metrics[rawMetric.Name]
	if !ok {
		return nil, errcodes.ErrMetricNotFound
	}

	if metric.MType != rawMetric.MType {
		return nil, errcodes.ErrMetricNotFound
	}

	return metric, nil
}

func (m *MemStorage) GetMetrics(_ context.Context) ([]types.Metric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make([]types.Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, utils.Deref(metric))
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
		m.metrics[metric.Name] = utils.Ptr(metric)
	}

	return nil
}
