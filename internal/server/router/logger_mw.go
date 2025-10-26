package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

// LoggerMiddleware добавляет zerolog в gin
func LoggerMiddleware(baseLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		l := baseLogger.With().
			Str("method", c.Request.Method).
			Str("url", c.Request.URL.Path).
			Str("req_id", uuid.NewString()[:8]).
			Logger()

		// Кладём логгер в обычный контекст
		ctx := context.WithValue(c.Request.Context(), logger.LogKey, &l)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		l = l.With().
			Int("status", status).
			Str("duration", time.Since(start).String()).
			Int("size", c.Writer.Size()).
			Logger()

		if len(c.Errors) > 0 || status >= 400 {
			errLog := l.Error()
			if status == http.StatusInternalServerError {
				errLog = errLog.Stack()
			}
			if len(c.Errors) > 0 {
				errLog.Err(c.Errors[0].Err)
			}
			errLog.Msg("Ошибка выполнения запроса")
			return
		}

		l.Info().Msg("Запрос успешно выполнен")
	}
}
