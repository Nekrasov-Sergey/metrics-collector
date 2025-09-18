package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) GetMetric(c *gin.Context) {
	ctx := c.Request.Context()

	metricName := types.MetricName(c.Param("name"))
	if metricName == "" {
		logger.Error(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

	metricTyp := types.MetricType(c.Param("type"))
	switch metricTyp {
	case types.Gauge, types.Counter:
	default:
		logger.Error(c, errors.Errorf("некорректный тип метрики: %s", metricTyp), http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(ctx, metricTyp, metricName)
	if err != nil {
		if errors.Is(err, errcodes.ErrMetricNotFound) {
			logger.Error(c, errcodes.ErrMetricNotFound, http.StatusNotFound)
			return
		}
		logger.Error(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, metric.Value)
}
