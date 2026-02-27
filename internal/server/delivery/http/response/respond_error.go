package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondError обрабатывает ошибку и формирует JSON-ответ
func RespondError(c *gin.Context, err error, status int) {
	if err == nil {
		c.AbortWithStatus(status)
		return
	}

	if status == http.StatusForbidden || status >= http.StatusInternalServerError {
		c.AbortWithStatusJSON(status, gin.H{
			"error": http.StatusText(status),
		})
	} else {
		c.AbortWithStatusJSON(status, gin.H{
			"error": err.Error(),
		})
	}

	_ = c.Error(err)
}
