package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func (h *Handler) ping(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.service.Ping(ctx); err != nil {
		logger.InternalServerError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
