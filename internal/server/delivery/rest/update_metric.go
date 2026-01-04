package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// UpdateMetric обновляет значение метрики на основе JSON-тела запроса.
//
// Ожидает JSON с полями:
//   - id  — имя метрики (обязательное)
//   - type — тип метрики: "gauge" или "counter" (обязательное)
//   - value — для gauge: новое значение float64
//   - delta — для counter: новое значение int64
//
// Пример запроса:
//
//	POST /value
//	{
//	  "id": "Alloc",
//	  "type": "gauge",
//	  "value": 123.45
//	}
//
// Возможные ответы:
//   - 200 OK — метрика успешно обновлена
//   - 400 Bad Request — некорректное тело запроса, отсутствие имени или значения, некорректный тип метрики
//   - 404 Not Found — имя метрики не указано
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) UpdateMetric(c *gin.Context) {
	ctx := c.Request.Context()

	metric := &types.Metric{}
	if err := c.ShouldBindJSON(metric); err != nil {
		logger.RespondError(c, errors.Wrap(err, "не удалось распарсить тело запроса"), http.StatusBadRequest)
		return
	}

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

	if err := h.service.UpdateMetric(ctx, metric); err != nil {
		logger.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
