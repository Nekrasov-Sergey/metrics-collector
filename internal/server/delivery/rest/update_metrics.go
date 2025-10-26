package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) updateMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	metrics := make([]types.Metric, 0)
	if err := c.ShouldBindJSON(&metrics); err != nil {
		logger.Error(c, errors.Wrap(err, "не удалось распарсить тело запроса"), http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if metric.Name == "" {
			logger.Error(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
			return
		}

		switch metric.MType {
		case types.Gauge:
			if metric.Value == nil {
				logger.Error(c, errors.New("значение метрики Gauge не задано"), http.StatusBadRequest)
				return
			}
		case types.Counter:
			if metric.Delta == nil {
				logger.Error(c, errors.New("значение метрики Counter не задано"), http.StatusBadRequest)
				return
			}
		default:
			logger.Error(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
			return
		}
	}

	if err := h.service.UpdateMetrics(ctx, metrics); err != nil {
		logger.InternalServerError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
