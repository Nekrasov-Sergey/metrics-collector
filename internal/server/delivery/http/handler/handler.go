package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

// Service описывает бизнес-логику работы с метриками.
//
// Интерфейс используется HTTP-обработчиками (Handler) и инкапсулирует
// операции получения, обновления и хранения метрик, а также проверку доступности зависимостей сервиса.
type Service interface {
	// UpdateMetric обновляет одну метрику.
	UpdateMetric(ctx context.Context, metric *types.Metric) error

	// GetMetric возвращает метрику по имени и типу.
	GetMetric(ctx context.Context, rawMetric *types.Metric) (metric *types.Metric, err error)

	// GetMetrics возвращает список всех метрик.
	GetMetrics(ctx context.Context) (metrics []types.Metric, err error)

	// SaveMetricsToFile сохраняет текущее состояние метрик в файл.
	// Метод не возвращает ошибку и предназначен для фонового использования.
	SaveMetricsToFile(ctx context.Context)

	// Ping проверяет доступность зависимостей сервиса (например, базы данных).
	Ping(ctx context.Context) error

	// UpdateMetrics обновляет несколько метрик.
	UpdateMetrics(ctx context.Context, metrics []types.Metric) error
}

// Audit описывает интерфейс аудита событий сервиса.
//
// Используется Handler для регистрации значимых событий,
// связанных с работой сервиса.
//
//go:generate minimock -i Audit -o ./mocks/audit.go -n AuditMock
type Audit interface {
	// Info регистрирует информационное событие аудита.
	Info(ctx context.Context, event *types.AuditEvent)
}

type Option func(*Handler)

func WithStoreInterval(storeInterval config.SecondDuration) Option {
	return func(h *Handler) {
		h.storeInterval = storeInterval
	}
}

// Handler обрабатывает HTTP-запросы сервиса метрик.
//
// Использует слой бизнес-логики (Service) и аудит событий (Audit).
type Handler struct {
	service       Service
	audit         Audit
	logger        zerolog.Logger
	storeInterval config.SecondDuration
}

func New(service Service, audit Audit, logger zerolog.Logger, opts ...Option) *Handler {
	h := &Handler{
		service: service,
		audit:   audit,
		logger:  logger,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/value/:type/:name", h.GetMetricByPath)
	r.POST("/value/", h.GetMetric)
	r.GET("/", h.GetMetrics)

	r.POST("/update/:type/:name/:value", h.UpdateMetricByPath)
	r.POST("/update/", h.UpdateMetric)
	r.POST("/updates", h.UpdateMetrics)

	r.GET("/ping", h.Ping)
}
