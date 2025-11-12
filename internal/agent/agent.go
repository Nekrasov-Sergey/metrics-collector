package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/common"
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

	metricsChan := make(chan []types.Metric, 1000)

	producersWG := sync.WaitGroup{}

	producersWG.Add(1)
	go func() {
		defer producersWG.Done()

		pollTicker := time.NewTicker(time.Duration(a.config.PollInterval))
		defer pollTicker.Stop()

		a.Poll(ctx, metricsChan, pollTicker)
	}()

	producersWG.Add(1)
	go func() {
		defer producersWG.Done()

		pollTicker := time.NewTicker(time.Duration(a.config.PollInterval))
		defer pollTicker.Stop()

		a.AdditionalPoll(ctx, metricsChan, pollTicker)
	}()

	workersCount := a.config.RateLimit
	if workersCount <= 0 {
		workersCount = 1
	}

	workersWG := sync.WaitGroup{}

	for w := range workersCount {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()

			reportTicker := time.NewTicker(time.Duration(a.config.ReportInterval))
			defer reportTicker.Stop()

			a.Report(ctx, metricsChan, reportTicker, w+1)
		}()
	}

	producersWG.Wait()
	close(metricsChan)
	workersWG.Wait()
	a.logger.Info().Msg("Агент остановлен")

	return nil
}

func (a *Agent) Poll(ctx context.Context, metricsChan chan<- []types.Metric, pollTicker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			memStats := &runtime.MemStats{}
			runtime.ReadMemStats(memStats)

			metrics := make([]types.Metric, 0)

			for name, f := range getGaugeMetrics() {
				metrics = append(metrics, types.Metric{
					Name:  name,
					MType: types.Gauge,
					Value: f(memStats),
				})
			}
			metrics = append(metrics, types.Metric{
				Name:  types.RandomValue,
				MType: types.Gauge,
				Value: utils.Ptr(rand.Float64()),
			})
			metrics = append(metrics, types.Metric{
				Name:  types.PollCount,
				MType: types.Counter,
				Delta: utils.Ptr(int64(1)),
			})

			metricsChan <- metrics
			a.logger.Info().Msg("Метрики собраны")
		}
	}
}

func (a *Agent) AdditionalPoll(ctx context.Context, metricsChan chan<- []types.Metric, pollTicker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			vmStat, err := mem.VirtualMemory()
			if err != nil {
				a.logger.Error().Err(err).Msg("Ошибка получения памяти")
				continue
			}

			// Получаем загрузку CPU по каждому ядру
			cpuUsages, err := cpu.Percent(0, true)
			if err != nil {
				a.logger.Error().Err(err).Msg("Ошибка получения CPU")
				continue
			}

			metrics := []types.Metric{
				{
					Name:  types.TotalMemory,
					MType: types.Gauge,
					Value: utils.Ptr(float64(vmStat.Total) / 1024 / 1024),
				},
				{
					Name:  types.FreeMemory,
					MType: types.Gauge,
					Value: utils.Ptr(float64(vmStat.Available) / 1024 / 1024),
				},
				{
					Name:  types.PollCount,
					MType: types.Counter,
					Delta: utils.Ptr(int64(1)),
				},
			}

			for i, usage := range cpuUsages {
				metrics = append(metrics, types.Metric{
					Name:  types.MetricName(fmt.Sprintf("CPUutilization%d", i+1)),
					MType: types.Gauge,
					Value: utils.Ptr(math.Round(usage*100) / 100),
				})
			}

			metricsChan <- metrics
			a.logger.Info().Msg("Дополнительные метрики собраны")
		}
	}
}

func (a *Agent) Report(ctx context.Context, metricsChan <-chan []types.Metric, reportTicker *time.Ticker, w int) {
	a.logger.Info().Int("воркер", w).Msg("Запущен воркер")

loop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTicker.C:
			for {
				select {
				case <-ctx.Done():
					return
				case metrics, ok := <-metricsChan:
					if !ok {
						return
					}
					a.sendMetrics(ctx, metrics, w)
				default:
					continue loop // канал опустел
				}
			}
		}
	}
}

func (a *Agent) sendMetrics(ctx context.Context, metrics []types.Metric, w int) {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось спарсить метрики в json")
		return
	}

	compressedMetrics, err := a.getCompressedMetrics(metricsJSON)
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось сжать метрики")
		return
	}

	path, err := url.JoinPath("http://", a.config.Addr, "/updates")
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось сформировать url")
		return
	}

	req := a.client.R().
		SetContext(ctx).
		SetBody(compressedMetrics).
		SetHeader("Content-Encoding", "gzip")

	if a.config.Key != "" {
		hashSHA256 := common.HMACSHA256([]byte(a.config.Key), metricsJSON)
		req.SetHeader("HashSHA256", hashSHA256)
	}

	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}
	for i, delay := range delays {
		resp, err := req.Post(path)
		if err == nil {
			if resp.IsError() {
				a.logger.Error().Int("воркер", w).Msgf("Сервер вернул ошибку: %s", resp.String())
				return
			}
			a.logger.Info().Int("воркер", w).Msgf("Отправлены метрики на сервер %s", a.config.Addr)
			return
		}

		var urlErr *url.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			a.logger.Error().Err(err).Int("воркер", w).Msg("Истек таймаут запроса")
			continue
		case errors.Is(err, context.Canceled):
			a.logger.Error().Err(err).Int("воркер", w).Msg("Контекст отменен")
			continue
		case errors.As(err, &urlErr):
			a.logger.Error().Err(err).Int("воркер", w).Msgf("Сервер недоступен, попытка №%d", i+1)
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					a.logger.Error().Int("воркер", w).Msg("Запрос отменён контекстом во время ожидания")
					continue
				case <-timer.C:
				}
			}
		default:
			a.logger.Error().Err(err).Int("воркер", w).Msg("Неизвестная ошибка")
			continue
		}
	}

	a.logger.Error().Int("воркер", w).Msg("Все попытки отправки исчерпаны")
}

func (a *Agent) getCompressedMetrics(metricsJSON []byte) ([]byte, error) {
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)

	if _, err := zw.Write(metricsJSON); err != nil {
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
