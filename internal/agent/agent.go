package agent

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Agent struct {
	client *resty.Client
	config *agentconfig.Config
	logger zerolog.Logger
}

func New(client *resty.Client, config *agentconfig.Config, logger zerolog.Logger) *Agent {
	return &Agent{
		client: client,
		config: config,
		logger: logger,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info().Msg("Запущен агент для сбора метрик")

	pollTicker := time.NewTicker(time.Duration(a.config.PollInterval))
	reportTicker := time.NewTicker(time.Duration(a.config.ReportInterval))
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	metrics := make(map[types.MetricName]types.Metric)
	var pollCount float64

	for {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Агент остановлен")
			return nil
		case <-pollTicker.C:
			pollCount = a.Poll(metrics, pollCount)
			a.logger.Info().Msg("Метрики собраны")
		case <-reportTicker.C:
			isSuccess := a.Report(metrics)
			if isSuccess {
				a.logger.Info().Msgf("Отправлено %d метрик на сервер %s", len(metrics), a.config.Addr)
				pollCount = 0
			}
		}
	}
}

func (a *Agent) Poll(metrics map[types.MetricName]types.Metric, pollCount float64) float64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	for name, f := range getGaugeMetrics() {
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

	return pollCount
}

func (a *Agent) Report(metrics map[types.MetricName]types.Metric) bool {
	isSuccess := true
	for _, metric := range metrics {
		_, err := a.client.R().
			SetPathParams(map[string]string{
				"type":  string(metric.Type),
				"name":  string(metric.Name),
				"value": strconv.FormatFloat(metric.Value, 'f', -1, 64),
			}).
			Post(a.config.Addr + "/update/{type}/{name}/{value}")
		if err != nil {
			a.logger.Error().Err(err).Send()
			isSuccess = false
			break
		}
	}
	return isSuccess
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
