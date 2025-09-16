package memstorage

import (
	"context"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (m *MemStorage) UpdateGaugeMetric(ctx context.Context, metricName string, gaugeValue float64) {
	m.Lock()
	defer m.Unlock()
	m.GaugeMetrics[metricName] = gaugeValue
	logger.C(ctx).Info().
		Str("тип", string(types.Gauge)).
		Str("имя", metricName).
		Float64("значение", m.GaugeMetrics[metricName]).
		Msg("Обновлённая метрика")
}

func (m *MemStorage) UpdateCounterMetric(ctx context.Context, metricName string, counterValue int64) {
	m.Lock()
	defer m.Unlock()
	m.CounterMetrics[metricName] += counterValue
	logger.C(ctx).Info().
		Str("тип", string(types.Counter)).
		Str("имя", metricName).
		Int64("значение", m.CounterMetrics[metricName]).
		Msg("Обновлённая метрика")
}
