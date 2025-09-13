package handler

import (
	"context"

	"github.com/gin-gonic/gin"
)

type RepoInterface interface {
	UpdateGaugeMetric(ctx context.Context, metricName string, gaugeValue float64)
	UpdateCounterMetric(ctx context.Context, metricName string, counterValue int64)
}

type Handler struct {
	repo RepoInterface
}

func New(repo RepoInterface) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/update/:type/:name/:value", h.UpdateMetric)
}
