package handler

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

func (h *Handler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := h.service.GetMetrics(ctx)
	if err != nil {
		logger.Error(c, err, http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("metrics").Parse(metricsHTML)
	if err != nil {
		logger.Error(c, errors.WithStack(err), http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(c.Writer, metrics)
	if err != nil {
		logger.Error(c, errors.WithStack(err), http.StatusInternalServerError)
		return
	}
}
