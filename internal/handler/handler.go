package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Service interface {
	UpdateMetric(ctx context.Context, typ types.MetricType, name types.MetricName, value float64) error
	GetMetric(ctx context.Context, typ types.MetricType, name types.MetricName) (metric types.Metric, err error)
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
	r.POST("/update/:type/:name/:value", h.UpdateMetric)
	r.GET("/value/:type/:name", h.GetMetric)
	r.GET("/", h.GetMetrics)
}
