package router

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// CheckIPMiddleware проверяет, что IP-адрес агента входит в доверенную подсеть
func CheckIPMiddleware(trustedNet *net.IPNet) gin.HandlerFunc {
	return func(c *gin.Context) {
		if trustedNet == nil {
			c.Next()
			return
		}

		ipStr := c.GetHeader("X-Real-IP")
		if ipStr == "" {
			logger.RespondError(c, errors.New("отсутствует заголовок X-Real-IP"), http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			logger.RespondError(c, errors.New("некорректный IP-адрес"), http.StatusForbidden)
			return
		}

		if !trustedNet.Contains(ip) {
			logger.RespondError(c, errors.New("IP не входит в CIDR"), http.StatusForbidden)
			return
		}

		c.Next()
	}
}
