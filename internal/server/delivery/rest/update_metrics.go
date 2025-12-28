package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) updateMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	var metrics []types.Metric
	if err := c.ShouldBindJSON(&metrics); err != nil {
		logger.RespondError(c, errors.Wrap(err, "не удалось распарсить метрики"), http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if metric.Name == "" {
			logger.RespondError(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
			return
		}

		switch metric.MType {
		case types.Gauge:
			if metric.Value == nil {
				logger.RespondError(c, errors.New("значение метрики Gauge не задано"), http.StatusBadRequest)
				return
			}
		case types.Counter:
			if metric.Delta == nil {
				logger.RespondError(c, errors.New("значение метрики Counter не задано"), http.StatusBadRequest)
				return
			}
		default:
			logger.RespondError(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
			return
		}
	}

	if err := h.service.UpdateMetrics(ctx, metrics); err != nil {
		logger.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	metricNames := make([]types.MetricName, 0, len(metrics))
	for _, metric := range metrics {
		metricNames = append(metricNames, metric.Name)
	}

	event := &types.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: c.RemoteIP(),
	}

	h.audit.Info(ctx, event)

	c.Status(http.StatusOK)
}
