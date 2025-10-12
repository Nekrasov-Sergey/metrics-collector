package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (h *Handler) updateMetricOld(c *gin.Context) {
	ctx := c.Request.Context()

	var metric types.Metric

	metric.Name = types.MetricName(c.Param("name"))
	if metric.Name == "" {
		logger.Error(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

	metric.MType = types.MetricType(c.Param("type"))
	switch metric.MType {
	case types.Gauge:
		value, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			logger.Error(c, errors.Wrap(err, "значение метрики не float64"), http.StatusBadRequest)
			return
		}
		metric.Value = utils.Ptr(value)
	case types.Counter:
		value, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			logger.Error(c, errors.Wrap(err, "значение метрики не int64"), http.StatusBadRequest)
			return
		}
		metric.Delta = utils.Ptr(value)
	default:
		logger.Error(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMetric(ctx, metric); err != nil {
		logger.InternalServerError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
