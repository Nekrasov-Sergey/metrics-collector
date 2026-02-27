package router

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/response"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/cryptoutil"
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
				response.RespondError(c, errors.Wrap(err, "не удалось прочитать тело запроса"), http.StatusInternalServerError)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			expectedHex := c.GetHeader("HashSHA256")
			isValid := cryptoutil.VerifyHMACSHA256([]byte(key), bodyBytes, expectedHex)
			if !isValid {
				response.RespondError(c, errors.New("Хеш запроса недействителен"), http.StatusBadRequest)
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
			respHash := cryptoutil.HMACSHA256([]byte(key), bw.body.Bytes())
			c.Header("HashSHA256", respHash)
		}
	}
}
