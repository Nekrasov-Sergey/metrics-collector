package rest

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

//go:embed static/metrics.html
var metricsHTML string

// GetMetrics возвращает HTML-страницу со списком всех метрик.
//
// Пример запроса:
//
//	GET /
//
// Возможные ответы:
//   - 200 OK — HTML-страница со списком метрик успешно сформирована
//   - 500 Internal Server Error — внутренняя ошибка сервиса
func (h *Handler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := h.service.GetMetrics(ctx)
	if err != nil {
		logger.RespondError(c, err, http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("metrics").Parse(metricsHTML)
	if err != nil {
		logger.RespondError(c, errors.WithStack(err), http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, metrics); err != nil {
		logger.RespondError(c, errors.WithStack(err), http.StatusInternalServerError)
		return
	}
}
