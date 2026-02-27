package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
)

// GetMetric возвращает значение метрики по имени и типу.
//
// Ожидает JSON-тело запроса со следующими полями:
//   - id — имя метрики
//   - type — тип метрики (gauge или counter)
//
// Пример запроса:
//
//	POST /value
//	{
//	  "id": "Alloc",
//	  "type": "gauge"
//	}
//
// Возможные ответы:
//   - 200 OK — метрика успешно найдена и возвращена
//   - 400 Bad Request — некорректное тело запроса или тип метрики
//   - 404 Not Found — имя метрики не указано или метрика не найдена
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) GetMetric(c *gin.Context) {
	ctx := c.Request.Context()

	metric := &types.Metric{}
	if err := c.ShouldBindJSON(metric); err != nil {
		response.RespondError(c, errors.Wrap(err, "не удалось распарсить тело запроса"), http.StatusBadRequest)
		return
	}

	if metric.Name == "" {
		response.RespondError(c, errors.New("отсутствует имя метрики"), http.StatusNotFound)
		return
	}

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

	c.JSON(http.StatusOK, metric)
}
