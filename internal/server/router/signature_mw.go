package router

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/common"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// SignatureMiddleware проверяет подпись запроса и добавляет подпись к ответу.
func SignatureMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("HashSHA256") != "" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.InternalServerError(c, errors.Wrap(err, "не удалось прочитать тело запроса"))
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			expectedHex := c.GetHeader("HashSHA256")
			isValid := common.VerifyHMACSHA256([]byte(key), bodyBytes, expectedHex)
			if !isValid {
				logger.Error(c, errors.New("Хеш запроса недействителен"), http.StatusBadRequest)
				return
			}
		}

		bw := &bodyWriter{
			body:           bytes.NewBuffer(nil),
			ResponseWriter: c.Writer,
		}
		c.Writer = bw

		c.Next()

		if key != "" {
			respHash := common.HMACSHA256([]byte(key), bw.body.Bytes())
			c.Header("HashSHA256", respHash)
		}
	}
}
