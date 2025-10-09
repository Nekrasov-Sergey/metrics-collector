package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Service interface {
	UpdateMetric(ctx context.Context, metric types.Metric) error
	GetMetric(ctx context.Context, rowMetric types.Metric) (metric types.Metric, err error)
	GetMetrics(ctx context.Context) (metrics []types.Metric, err error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/update/:type/:name/:value", h.UpdateMetricOld)
	r.GET("/value/:type/:name", h.GetMetricOld)
	r.GET("/", h.GetMetrics)

	r.POST("/update", h.UpdateMetric)
	r.POST("/value", h.GetMetric)
}
