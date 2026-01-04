package service

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

// Repository описывает интерфейс хранилища метрик.
//
// Инкапсулирует операции получения, обновления и хранения метрик,
// а также управление соединением с хранилищем данных.
//
//go:generate minimock -i github.com/Nekrasov-Sergey/metrics-collector/internal/server/service.Repository -o ./mocks/repo.go -n RepoMock
type Repository interface {
	// UpdateMetric обновляет одну метрику в хранилище.
	UpdateMetric(ctx context.Context, metric *types.Metric) error

	// GetMetric возвращает метрику по имени и типу.
	GetMetric(ctx context.Context, rawMetric *types.Metric) (*types.Metric, error)

	// GetMetrics возвращает все метрики.
	GetMetrics(ctx context.Context) ([]types.Metric, error)

	// UpdateMetrics обновляет несколько метрик.
	UpdateMetrics(ctx context.Context, metrics []types.Metric) error

	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error

	// Close закрывает соединение с хранилищем и освобождает ресурсы.
	Close() error
}

type Option func(*Service)

func WithStoreInterval(storeInterval config.SecondDuration) Option {
	return func(s *Service) {
		s.storeInterval = storeInterval
	}
}

func WithRestore(restore bool) Option {
	return func(s *Service) {
		s.restore = restore
	}
}

func WithFileStoragePath(fileStoragePath string) Option {
	return func(s *Service) {
		s.fileStoragePath = fileStoragePath
	}
}

// Service реализует бизнес-логику работы с метриками.
//
// Отвечает за получение, обновление и сохранение метрик.
// Использует слой репозитория (Repository).
type Service struct {
	repo            Repository
	logger          zerolog.Logger
	storeInterval   config.SecondDuration
	restore         bool
	fileStoragePath string
}

func New(ctx context.Context, repo Repository, logger zerolog.Logger, opts ...Option) *Service {
	s := &Service{
		repo:   repo,
		logger: logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.loadMetricsFromFile(ctx)

	return s
}
