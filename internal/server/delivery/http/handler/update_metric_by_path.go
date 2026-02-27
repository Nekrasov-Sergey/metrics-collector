package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

// UpdateMetricByPath обновляет значение метрики по имени, типу и значению, переданным в параметрах пути.
//
// Ожидает следующие path-параметры:
//   - type  — тип метрики (gauge или counter)
//   - name  — имя метрики
//   - value — новое значение метрики
//
// Пример запроса:
//
//	POST /update/gauge/Alloc/123.45
//
// Поведение по типу метрики:
//   - gauge   — значение value парсится как float64
//   - counter — значение value парсится как int64
//
// Возможные ответы:
//   - 200 OK — метрика успешно обновлена
//   - 400 Bad Request — некорректный тип метрики или значение value
//   - 404 Not Found — имя метрики не указано
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) UpdateMetricByPath(c *gin.Context) {
	ctx := c.Request.Context()

	metric := &types.Metric{}

	metric.Name = types.MetricName(c.Param("name"))
	if metric.Name == "" {
		response.RespondError(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

	metric.MType = types.MetricType(c.Param("type"))
	switch metric.MType {
	case types.Gauge:
		value, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			response.RespondError(c, errors.Wrap(err, "значение метрики не float64"), http.StatusBadRequest)
			return
		}
		metric.Value = utils.Ptr(value)
	case types.Counter:
		value, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			response.RespondError(c, errors.Wrap(err, "значение метрики не int64"), http.StatusBadRequest)
			return
		}
		metric.Delta = utils.Ptr(value)
	default:
		response.RespondError(c, errors.Errorf("некорректный тип метрики: %s", metric.MType), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMetric(ctx, metric); err != nil {
		response.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
