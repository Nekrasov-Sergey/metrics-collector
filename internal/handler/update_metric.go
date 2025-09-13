package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) UpdateMetric(c *gin.Context) {
	ctx := c.Request.Context()

	metricName := c.Param("name")
	if metricName == "" {
		logger.Error(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

	metricTyp := types.MetricType(c.Param("type"))
	switch metricTyp {
	case types.Gauge:
		gaugeValue, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			logger.Error(c, errors.Wrap(err, "значение метрики не float64"), http.StatusBadRequest)
			return
		}
		h.repo.UpdateGaugeMetric(ctx, metricName, gaugeValue)
	case types.Counter:
		counterValue, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			logger.Error(c, errors.Wrap(err, "значение метрики не int64"), http.StatusBadRequest)
			return
		}
		h.repo.UpdateCounterMetric(ctx, metricName, counterValue)
	default:
		logger.Error(c, errors.Errorf("некорректный тип метрики: %s", metricTyp), http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}
