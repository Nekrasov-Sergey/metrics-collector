package router

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// DecryptMiddleware расшифровывает тело запроса с помощью rsa
func DecryptMiddleware(cryptoKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cryptoKey == "" {
			c.Next()
			return
		}

		privateKeyBytes, err := os.ReadFile(cryptoKey)
		if err != nil {
			logger.RespondError(c, errors.New("не удалось прочитать файл с приватным ключом"), http.StatusInternalServerError)
			return
		}

		privateKeyPemBlock, _ := pem.Decode(privateKeyBytes)
		if privateKeyPemBlock == nil {
			logger.RespondError(c, errors.New("не удалось декодировать pem-блок приватного ключа"), http.StatusInternalServerError)
			return
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
		if err != nil {
			logger.RespondError(c, errors.New("не удалось распарсить rsa приватный ключ"), http.StatusInternalServerError)
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
