package memstorage

import (
	"context"
	"sort"

	"github.com/Nekrasov-Sergey/metrics-collector/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (m *MemStorage) UpdateMetric(ctx context.Context, typ types.MetricType, name types.MetricName, value float64) error {
	m.Lock()
	defer m.Unlock()

	m.Metrics[name] = types.Metric{
		Name:  name,
		Type:  typ,
		Value: value,
	}
	logger.C(ctx).Info().
		Str("тип", string(typ)).
		Str("имя", string(name)).
		Any("значение", value).
		Msg("Обновлённая метрика")

	return nil
}

func (m *MemStorage) GetMetric(_ context.Context, typ types.MetricType, name types.MetricName) (metric types.Metric, err error) {
	m.RLock()
	defer m.RUnlock()

	metric, ok := m.Metrics[name]
	if !ok {
		return types.Metric{}, errcodes.ErrMetricNotFound
	}

	if metric.Type != typ {
		return types.Metric{}, errcodes.ErrMetricNotFound
	}

	return metric, nil
}

func (m *MemStorage) GetMetrics(_ context.Context) (metrics []types.Metric, err error) {
	m.RLock()
	defer m.RUnlock()

	metrics = make([]types.Metric, 0, len(m.Metrics))
	for _, metric := range m.Metrics {
		metrics = append(metrics, metric)
	}
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics, nil
}
