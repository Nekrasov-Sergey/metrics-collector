package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) getMetric(c *gin.Context) {
	ctx := c.Request.Context()

	var metric types.Metric
	if err := c.ShouldBindJSON(&metric); err != nil {
		logger.Error(c, errors.Wrap(err, "не удалось распарсить тело запроса"), http.StatusBadRequest)
		return
	}

	if metric.Name == "" {
		logger.Error(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

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

	c.JSON(http.StatusOK, metric)
}
