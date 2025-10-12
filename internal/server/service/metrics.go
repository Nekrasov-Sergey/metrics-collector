package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (s *Service) UpdateMetric(ctx context.Context, metric types.Metric) error {
	if metric.MType == types.Counter {
		counterMetric, err := s.GetMetric(ctx, metric)
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
			log.Info().Msg("Файл для загрузки метрик не существует")
			return
		}
		log.Error().Err(err).Msg("Не удалось прочитать файл с метриками")
		return
	}

	var metrics []types.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить метрики")
		return
	}

	if err := s.repo.UpdateMetrics(ctx, metrics); err != nil {
		log.Error().Err(err).Msg("Не удалось загрузить метрики в репозиторий")
		return
	}

	log.Info().Msg("Метрики успешно загружены")
}

func (s *Service) SaveMetricsToFile(ctx context.Context) {
	metrics, err := s.repo.GetMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось получить метрики из репозитория")
		return
	}

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		log.Error().Err(err).Msg("Не удалось спарсить метрики в json")
		return
	}

	if err := os.MkdirAll(filepath.Dir(s.config.FileStoragePath), 0755); err != nil {
		log.Error().Err(err).Msg("Не удалось открыть файл для записи метрик")
		return
	}

	if err := os.WriteFile(s.config.FileStoragePath, data, 0644); err != nil {
		log.Error().Err(err).Msg("Не удалось записать метрики в файл")
		return
	}

	log.Info().Msg("Метрики сохранены в файл")
}
