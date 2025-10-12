package rest

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Service interface {
	UpdateMetric(ctx context.Context, metric types.Metric) error
	GetMetric(ctx context.Context, rowMetric types.Metric) (metric types.Metric, err error)
	GetMetrics(ctx context.Context) (metrics []types.Metric, err error)
	SaveMetricsToFile(ctx context.Context)
}

type Handler struct {
	service Service
	config  *serverconfig.Config
}

func New(service Service, config *serverconfig.Config) *Handler {
	return &Handler{
		service: service,
		config:  config,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/update/:type/:name/:value", h.updateMetricOld)
	r.GET("/value/:type/:name", h.getMetricOld)
	r.GET("/", h.getMetrics)

	r.POST("/update", h.updateMetric)
	r.POST("/value", h.getMetric)
}
