package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Service interface {
	UpdateMetric(ctx context.Context, metric *types.Metric) error
	GetMetric(ctx context.Context, rawMetric *types.Metric) (metric *types.Metric, err error)
	GetMetrics(ctx context.Context) (metrics []types.Metric, err error)
	SaveMetricsToFile(ctx context.Context)
	Ping(ctx context.Context) error
	UpdateMetrics(ctx context.Context, metrics []types.Metric) error
}

//go:generate minimock -i github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest.Audit -o ./mocks/audit.go -n AuditMock
type Audit interface {
	Info(ctx context.Context, event *types.AuditEvent)
}

type Handler struct {
	service Service
	config  *config.ServerConfig
	logger  zerolog.Logger
	audit   Audit
}

func New(service Service, config *config.ServerConfig, logger zerolog.Logger, audit Audit) *Handler {
	return &Handler{
		service: service,
		config:  config,
		logger:  logger,
		audit:   audit,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/update/:type/:name/:value", h.updateMetricOld)
	r.GET("/value/:type/:name", h.getMetricOld)
	r.GET("/", h.getMetrics)

	r.POST("/update/", h.updateMetric)
	r.POST("/value/", h.getMetric)

	r.GET("/ping", h.ping)
	r.POST("/updates", h.updateMetrics)
}
