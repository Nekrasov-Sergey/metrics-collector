package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// UpdateMetrics обновляет несколько метрик одновременно на основе JSON-массива.
//
// Ожидает JSON-массив объектов с полями:
//   - id  — имя метрики (обязательное)
//   - type — тип метрики: "gauge" или "counter" (обязательное)
//   - value — для gauge: новое значение float64
//   - delta — для counter: новое значение int64
//
// Пример запроса:
//
//	POST /updates
//	[
//	  {
//	    "id": "Alloc",
//	    "type": "gauge",
//	    "value": 123.45
//	  },
//	  {
//	    "id": "PollCount",
//	    "type": "counter",
//	    "delta": 5
//	  }
//	]
//
// После успешного обновления создается событие аудита с именами обновленных метрик и IP клиента.
//
// Возможные ответы:
//   - 200 OK — все метрики успешно обновлены
//   - 400 Bad Request — некорректное тело запроса, отсутствие имени или значения, некорректный тип метрики
//   - 404 Not Found — имя одной из метрик не указано
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) UpdateMetrics(c *gin.Context) {
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
