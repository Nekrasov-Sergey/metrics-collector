package agent

import (
	"context"
	"fmt"
	"math"
	mrand "math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

// Run запускает основной цикл работы агента.
//
// Метод стартует горутины сбора метрик и воркеры их отправки.
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

// Poll периодически собирает базовые runtime-метрики.
//
// Сбор выполняется по тикеру и результаты отправляются в канал metricsChan.
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
				Value: utils.Ptr(mrand.Float64()),
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

// AdditionalPoll периодически собирает дополнительные системные метрики.
//
// Включает информацию о памяти, загрузке CPU и другие метрики операционной системы.
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

// Report запускает воркер отправки метрик на сервер.
//
// Воркер периодически читает накопленные метрики из канала и отправляет их на сервер с учетом ограничений и ретраев.
func (a *Agent) Report(ctx context.Context, metricsChan <-chan []types.Metric, reportTicker *time.Ticker, w int) {
	a.logger.Info().Int("воркер", w).Msg("Запущен воркер")

	for {
		select {
		case <-ctx.Done():
			a.flushOnShutdown(metricsChan, w)
			return
		case <-reportTicker.C:
			a.drain(ctx, metricsChan, w)
		}
	}
}

func (a *Agent) drain(ctx context.Context, metricsChan <-chan []types.Metric, w int) {
	for {
		select {
		case <-ctx.Done():
			return // контекст отменен
		case metrics, ok := <-metricsChan:
			if !ok {
				return // канал закрыт и пустой
			}
			a.sendMetrics(ctx, metrics, w)
		default:
			return // канал пустой
		}
	}
}

func (a *Agent) flushOnShutdown(metricsChan <-chan []types.Metric, w int) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case metrics, ok := <-metricsChan:
			if !ok {
				return // канал закрыт и пустой
			}
			a.sendMetrics(shutdownCtx, metrics, w)
		default:
			return // канал пустой
		}
	}
}

func (a *Agent) sendMetrics(ctx context.Context, metrics []types.Metric, w int) {
	if a.config.GRPCAddr != "" {
		a.gRPCSendMetrics(ctx, metrics, w)
	} else {
		a.httpSendMetrics(ctx, metrics, w)
	}
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
