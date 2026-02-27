package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
)

// GetMetricByPath возвращает значение метрики по имени и типу.
//
// Ожидает следующие path-параметры:
//   - name — имя метрики
//   - type — тип метрики (gauge или counter)
//
// Пример запроса:
//
//	GET /value/gauge/Alloc
//
// Возможные ответы:
//   - 200 OK — метрика успешно найдена
//   - 400 Bad Request — некорректный тип метрики
//   - 404 Not Found — метрика не найдена
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) GetMetricByPath(c *gin.Context) {
	ctx := c.Request.Context()

	metric := &types.Metric{}

	metric.Name = types.MetricName(c.Param("name"))

	metric.MType = types.MetricType(c.Param("type"))
	if !metric.MType.IsValid() {
		response.RespondError(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(ctx, metric)
	if err != nil {
		if errors.Is(err, errcodes.ErrMetricNotFound) {
			response.RespondError(c, errcodes.ErrMetricNotFound, http.StatusNotFound)
			return
		}
		response.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	if metric.MType == types.Gauge {
		c.JSON(http.StatusOK, metric.Value)
		return
	}
	c.JSON(http.StatusOK, metric.Delta)
}
