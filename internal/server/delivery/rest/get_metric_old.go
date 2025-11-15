package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) getMetricOld(c *gin.Context) {
	ctx := c.Request.Context()

	var metric types.Metric

	metric.Name = types.MetricName(c.Param("name"))

	metric.MType = types.MetricType(c.Param("type"))
	if !metric.MType.IsValid() {
		logger.Error(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(ctx, metric)
	if err != nil {
		if errors.Is(err, errcodes.ErrMetricNotFound) {
			logger.Error(c, errcodes.ErrMetricNotFound, http.StatusNotFound)
			return
		}
		logger.InternalServerError(c, err)
		return
	}

	if metric.MType == types.Gauge {
		c.JSON(http.StatusOK, metric.Value)
		return
	}
	c.JSON(http.StatusOK, metric.Delta)
}
