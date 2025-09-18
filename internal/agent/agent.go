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
	var pollCount float64

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
			pollCount = 0
		}
	}
}

type getMetricValue func(memStats *runtime.MemStats) float64

func getGaugeMetrics() map[types.MetricName]getMetricValue {
	return map[types.MetricName]getMetricValue{
		types.Alloc:         func(memStats *runtime.MemStats) float64 { return float64(memStats.Alloc) },
		types.BuckHashSys:   func(memStats *runtime.MemStats) float64 { return float64(memStats.BuckHashSys) },
		types.Frees:         func(memStats *runtime.MemStats) float64 { return float64(memStats.Frees) },
		types.GCCPUFraction: func(memStats *runtime.MemStats) float64 { return memStats.GCCPUFraction },
		types.GCSys:         func(memStats *runtime.MemStats) float64 { return float64(memStats.GCSys) },
		types.HeapAlloc:     func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapAlloc) },
		types.HeapIdle:      func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapIdle) },
		types.HeapInuse:     func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapInuse) },
		types.HeapObjects:   func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapObjects) },
		types.HeapReleased:  func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapReleased) },
		types.HeapSys:       func(memStats *runtime.MemStats) float64 { return float64(memStats.HeapSys) },
		types.LastGC:        func(memStats *runtime.MemStats) float64 { return float64(memStats.LastGC) },
		types.Lookups:       func(memStats *runtime.MemStats) float64 { return float64(memStats.Lookups) },
		types.MCacheInuse:   func(memStats *runtime.MemStats) float64 { return float64(memStats.MCacheInuse) },
		types.MCacheSys:     func(memStats *runtime.MemStats) float64 { return float64(memStats.MCacheSys) },
		types.MSpanInuse:    func(memStats *runtime.MemStats) float64 { return float64(memStats.MSpanInuse) },
		types.MSpanSys:      func(memStats *runtime.MemStats) float64 { return float64(memStats.MSpanSys) },
		types.Mallocs:       func(memStats *runtime.MemStats) float64 { return float64(memStats.Mallocs) },
		types.NextGC:        func(memStats *runtime.MemStats) float64 { return float64(memStats.NextGC) },
		types.NumForcedGC:   func(memStats *runtime.MemStats) float64 { return float64(memStats.NumForcedGC) },
		types.NumGC:         func(memStats *runtime.MemStats) float64 { return float64(memStats.NumGC) },
		types.OtherSys:      func(memStats *runtime.MemStats) float64 { return float64(memStats.OtherSys) },
		types.PauseTotalNs:  func(memStats *runtime.MemStats) float64 { return float64(memStats.PauseTotalNs) },
		types.StackInuse:    func(memStats *runtime.MemStats) float64 { return float64(memStats.StackInuse) },
		types.StackSys:      func(memStats *runtime.MemStats) float64 { return float64(memStats.StackSys) },
		types.Sys:           func(memStats *runtime.MemStats) float64 { return float64(memStats.Sys) },
		types.TotalAlloc:    func(memStats *runtime.MemStats) float64 { return float64(memStats.TotalAlloc) },
	}
}
