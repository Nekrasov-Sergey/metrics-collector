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
		logger.RespondError(c, errors.Wrap(err, "не удалось распарсить тело запроса"), http.StatusBadRequest)
		return
	}

	if metric.Name == "" {
		logger.RespondError(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

	if !metric.MType.IsValid() {
		logger.RespondError(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(ctx, metric)
	if err != nil {
		if errors.Is(err, errcodes.ErrMetricNotFound) {
			logger.RespondError(c, errcodes.ErrMetricNotFound, http.StatusNotFound)
			return
		}
		logger.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, metric)
}
