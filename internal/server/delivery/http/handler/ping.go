package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
)

// Ping проверяет доступность сервиса и соединение с базой данных.
//
// Пример запроса:
//
//	GET /ping
//
// Возможные ответы:
//   - 200 OK — соединение с базой данных успешно, сервис работает
//   - 500 Internal Server Error — внутренняя ошибка сервиса или недоступна база данных
func (h *Handler) Ping(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.service.Ping(ctx); err != nil {
		response.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
