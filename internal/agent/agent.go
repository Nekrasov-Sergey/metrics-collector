package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"math/rand/v2"
	"net/url"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

type Agent struct {
	config *agentconfig.Config
	client *resty.Client
	logger zerolog.Logger
}

func New(config *agentconfig.Config, client *resty.Client, logger zerolog.Logger) *Agent {
	return &Agent{
		config: config,
		client: client,
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
	var pollCount int64

	for {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Агент остановлен")
			return nil
		case <-pollTicker.C:
			pollCount = a.Poll(metrics, pollCount)
			a.logger.Info().Msg("Метрики собраны")
		case <-reportTicker.C:
			isSuccess := a.Report(ctx, metrics)
			if isSuccess {
				a.logger.Info().Msgf("Отправлено %d метрик на сервер %s", len(metrics), a.config.Addr)
				pollCount = 0
			}
		}
	}
}

func (a *Agent) Poll(metrics map[types.MetricName]types.Metric, pollCount int64) int64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	for name, f := range getGaugeMetrics() {
		metrics[name] = types.Metric{
			Name:  name,
			MType: types.Gauge,
			Value: f(memStats),
		}
	}
	metrics[types.RandomValue] = types.Metric{
		Name:  types.RandomValue,
		MType: types.Gauge,
		Value: utils.Ptr(rand.Float64()),
	}
	pollCount++
	metrics[types.PollCount] = types.Metric{
		Name:  types.PollCount,
		MType: types.Counter,
		Delta: utils.Ptr(pollCount),
	}

	return pollCount
}

func (a *Agent) Report(ctx context.Context, metricsMap map[types.MetricName]types.Metric) bool {
	if len(metricsMap) == 0 {
		return false
	}

	metrics := make([]types.Metric, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	compressedMetrics, err := a.getCompressedMetrics(metrics)
	if err != nil {
		a.logger.Error().Err(err).Msg("Не удалось сжать метрики")
		return false
	}

	path, err := url.JoinPath("http://", a.config.Addr, "/updates")
	if err != nil {
		a.logger.Error().Err(err).Msg("Не удалось сформировать url")
		return false
	}

	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}
	for i, delay := range delays {
		resp, err := a.client.R().
			SetContext(ctx).
			SetHeader("Content-Encoding", "gzip").
			SetBody(compressedMetrics).
			Post(path)

		if err == nil {
			if resp.IsError() {
				a.logger.Error().Msgf("Сервер вернул ошибку: %s", resp.String())
				return false
			}
			return true
		}

		var urlErr *url.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			a.logger.Error().Err(err).Msg("Истек таймаут запроса")
			return false
		case errors.Is(err, context.Canceled):
			a.logger.Error().Err(err).Msg("Контекст отменен")
			return false
		case errors.As(err, &urlErr):
			a.logger.Error().Err(err).Msgf("Сервер недоступен, попытка №%d", i+1)
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					a.logger.Error().Msg("Запрос отменён контекстом во время ожидания")
					return false
				case <-timer.C:
				}
			}
		default:
			a.logger.Error().Err(err).Msg("Неизвестная ошибка")
			return false
		}
	}

	a.logger.Error().Msg("Все попытки отправки исчерпаны")
	return false
}

func (a *Agent) getCompressedMetrics(metrics []types.Metric) ([]byte, error) {
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)

	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось спарсить метрики в json")
	}

	if _, err := zw.Write(data); err != nil {
		return nil, errors.Wrap(err, "не удалось записать данные для сжатия")
	}

	if err := zw.Close(); err != nil {
		return nil, errors.Wrap(err, "не удалось сжать данные")
	}

	return b.Bytes(), nil
}

type getMetricValue func(memStats *runtime.MemStats) *float64

func getGaugeMetrics() map[types.MetricName]getMetricValue {
	return map[types.MetricName]getMetricValue{
		types.Alloc:         func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.Alloc)) },
		types.BuckHashSys:   func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.BuckHashSys)) },
		types.Frees:         func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.Frees)) },
		types.GCCPUFraction: func(memStats *runtime.MemStats) *float64 { return utils.Ptr(memStats.GCCPUFraction) },
		types.GCSys:         func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.GCSys)) },
		types.HeapAlloc:     func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapAlloc)) },
		types.HeapIdle:      func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapIdle)) },
		types.HeapInuse:     func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapInuse)) },
		types.HeapObjects:   func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapObjects)) },
		types.HeapReleased:  func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapReleased)) },
		types.HeapSys:       func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.HeapSys)) },
		types.LastGC:        func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.LastGC)) },
		types.Lookups:       func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.Lookups)) },
		types.MCacheInuse:   func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.MCacheInuse)) },
		types.MCacheSys:     func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.MCacheSys)) },
		types.MSpanInuse:    func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.MSpanInuse)) },
		types.MSpanSys:      func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.MSpanSys)) },
		types.Mallocs:       func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.Mallocs)) },
		types.NextGC:        func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.NextGC)) },
		types.NumForcedGC:   func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.NumForcedGC)) },
		types.NumGC:         func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.NumGC)) },
		types.OtherSys:      func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.OtherSys)) },
		types.PauseTotalNs:  func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.PauseTotalNs)) },
		types.StackInuse:    func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.StackInuse)) },
		types.StackSys:      func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.StackSys)) },
		types.Sys:           func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.Sys)) },
		types.TotalAlloc:    func(memStats *runtime.MemStats) *float64 { return utils.Ptr(float64(memStats.TotalAlloc)) },
	}
}
