package router

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// DecryptMiddleware расшифровывает тело запроса с помощью rsa
func DecryptMiddleware(privateKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if privateKey == nil {
			c.Next()
			return
		}

		encryptedMessage, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.RespondError(c, errors.New("не удалось прочитать тело запроса"), http.StatusBadRequest)
			return
		}

		decryptedMessage, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedMessage)
		if err != nil {
			logger.RespondError(c, errors.New("не удалось расшифровать тело запроса"), http.StatusBadRequest)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(decryptedMessage))

		c.Next()
	}
}
