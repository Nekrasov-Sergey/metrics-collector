package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func Run(ctx context.Context) {
	l := logger.New()
	l.Info().Msg("Запущен агент для сбора метрик")

	memStats := &runtime.MemStats{}
	gaugeMetrics := getGaugeMetrics()

	client := resty.New()

	tickerPoll := time.NewTicker(types.PollInterval)
	defer tickerPoll.Stop()
	tickerReport := time.NewTicker(types.ReportInterval)
	defer tickerReport.Stop()

	metrics := make(map[types.MetricName]types.Metric)
	var pollCount int64

	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("Агент успешно остановлен")
			return
		case <-tickerPoll.C:
			runtime.ReadMemStats(memStats)
			for name, f := range gaugeMetrics {
				metrics[name] = types.Metric{
					Name:  name,
					Type:  types.Gauge,
					Value: f(memStats),
				}
			}
			metrics[types.RandomValue] = types.Metric{
				Name:  types.RandomValue,
				Type:  types.Gauge,
				Value: rand.Float64(),
			}
			pollCount++
			metrics[types.PollCount] = types.Metric{
				Name:  types.PollCount,
				Type:  types.Counter,
				Value: pollCount,
			}
			l.Info().Msg("Метрики успешно собраны")
		case <-tickerReport.C:
			fail := false
			for _, metric := range metrics {
				_, err := client.R().
					SetPathParams(map[string]string{
						"type":  string(metric.Type),
						"name":  string(metric.Name),
						"value": fmt.Sprintf("%v", metric.Value),
					}).
					Post("http://localhost:8080/update/{type}/{name}/{value}")
				if err != nil {
					l.Error().Err(err).Send()
					fail = true
					continue
				}
			}
			if !fail {
				l.Info().Msg("Метрики успешно отправлены")
			}
		}
	}
}

type getMetricValue func(memStats *runtime.MemStats) any

func getGaugeMetrics() map[types.MetricName]getMetricValue {
	return map[types.MetricName]getMetricValue{
		types.Alloc:         func(memStats *runtime.MemStats) any { return memStats.Alloc },
		types.BuckHashSys:   func(memStats *runtime.MemStats) any { return memStats.BuckHashSys },
		types.Frees:         func(memStats *runtime.MemStats) any { return memStats.Frees },
		types.GCCPUFraction: func(memStats *runtime.MemStats) any { return memStats.GCCPUFraction },
		types.GCSys:         func(memStats *runtime.MemStats) any { return memStats.GCSys },
		types.HeapAlloc:     func(memStats *runtime.MemStats) any { return memStats.HeapAlloc },
		types.HeapIdle:      func(memStats *runtime.MemStats) any { return memStats.HeapIdle },
		types.HeapInuse:     func(memStats *runtime.MemStats) any { return memStats.HeapInuse },
		types.HeapObjects:   func(memStats *runtime.MemStats) any { return memStats.HeapObjects },
		types.HeapReleased:  func(memStats *runtime.MemStats) any { return memStats.HeapReleased },
		types.HeapSys:       func(memStats *runtime.MemStats) any { return memStats.HeapSys },
		types.LastGC:        func(memStats *runtime.MemStats) any { return memStats.LastGC },
		types.Lookups:       func(memStats *runtime.MemStats) any { return memStats.Lookups },
		types.MCacheInuse:   func(memStats *runtime.MemStats) any { return memStats.MCacheInuse },
		types.MCacheSys:     func(memStats *runtime.MemStats) any { return memStats.MCacheSys },
		types.MSpanInuse:    func(memStats *runtime.MemStats) any { return memStats.MSpanInuse },
		types.MSpanSys:      func(memStats *runtime.MemStats) any { return memStats.MSpanSys },
		types.Mallocs:       func(memStats *runtime.MemStats) any { return memStats.Mallocs },
		types.NextGC:        func(memStats *runtime.MemStats) any { return memStats.NextGC },
		types.NumForcedGC:   func(memStats *runtime.MemStats) any { return memStats.NumForcedGC },
		types.NumGC:         func(memStats *runtime.MemStats) any { return memStats.NumGC },
		types.OtherSys:      func(memStats *runtime.MemStats) any { return memStats.OtherSys },
		types.PauseTotalNs:  func(memStats *runtime.MemStats) any { return memStats.PauseTotalNs },
		types.StackInuse:    func(memStats *runtime.MemStats) any { return memStats.StackInuse },
		types.StackSys:      func(memStats *runtime.MemStats) any { return memStats.StackSys },
		types.Sys:           func(memStats *runtime.MemStats) any { return memStats.Sys },
		types.TotalAlloc:    func(memStats *runtime.MemStats) any { return memStats.TotalAlloc },
	}
}
