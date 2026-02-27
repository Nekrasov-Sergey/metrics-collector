package router

import (
	"net"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
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
			response.RespondError(c, errors.New("отсутствует заголовок X-Real-IP"), http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			response.RespondError(c, errors.New("некорректный IP-адрес"), http.StatusForbidden)
			return
		}

		if !trustedNet.Contains(ip) {
			response.RespondError(c, errors.New("IP не входит в CIDR"), http.StatusForbidden)
			return
		}

		c.Next()
	}
}
