package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (s *Service) UpdateMetric(ctx context.Context, metric types.Metric) error {
	if metric.MType == types.Counter {
		counterMetric, err := s.repo.GetMetric(ctx, metric)
		if err != nil && !errors.Is(err, errcodes.ErrMetricNotFound) {
			return err
		}
		*metric.Delta += utils.Deref(counterMetric.Delta)
	}

	if err := s.repo.UpdateMetric(ctx, metric); err != nil {
		return err
	}

	if s.config.StoreInterval == 0 {
		s.SaveMetricsToFile(ctx)
	}

	return nil
}

func (s *Service) GetMetric(ctx context.Context, rowMetric types.Metric) (metric types.Metric, err error) {
	return s.repo.GetMetric(ctx, rowMetric)
}

func (s *Service) GetMetrics(ctx context.Context) (metrics []types.Metric, err error) {
	return s.repo.GetMetrics(ctx)
}

func (s *Service) loadMetricsFromFile(ctx context.Context) {
	if !s.config.Restore {
		return
	}

	data, err := os.ReadFile(s.config.FileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Info().Msg("Файл для загрузки метрик не существует")
			return
		}
		s.logger.Error().Err(err).Msg("Не удалось прочитать файл с метриками")
		return
	}

	var metrics []types.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		s.logger.Error().Err(err).Msg("Не удалось распарсить метрики")
		return
	}

	if err := s.repo.UpdateMetrics(ctx, metrics); err != nil {
		s.logger.Error().Err(err).Msg("Не удалось загрузить метрики в репозиторий")
		return
	}

	s.logger.Info().Msg("Метрики успешно загружены")
}

func (s *Service) SaveMetricsToFile(ctx context.Context) {
	metrics, err := s.repo.GetMetrics(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Не удалось получить метрики из репозитория")
		return
	}

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		s.logger.Error().Err(err).Msg("Не удалось спарсить метрики в json")
		return
	}

	if err := os.MkdirAll(filepath.Dir(s.config.FileStoragePath), 0755); err != nil {
		s.logger.Error().Err(err).Msg("Не удалось открыть файл для записи метрик")
		return
	}

	if err := os.WriteFile(s.config.FileStoragePath, data, 0644); err != nil {
		s.logger.Error().Err(err).Msg("Не удалось записать метрики в файл")
		return
	}

	s.logger.Info().Msg("Метрики сохранены в файл")
}

func (s *Service) PingDB(ctx context.Context) error {
	return s.repo.PingDB(ctx)
}
