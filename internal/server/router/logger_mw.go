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

		l = l.With().
			Int("status", c.Writer.Status()).
			Str("duration", time.Since(start).String()).
			Int("size", c.Writer.Size()).
			Logger()

		if len(c.Errors) > 0 {
			errLog := l.Error()
			if c.Writer.Status() == http.StatusInternalServerError {
				errLog = errLog.Stack()
			}
			errLog.Err(c.Errors[0].Err).Msg("Ошибка выполнения запроса")
			return
		}

		l.Info().Msg("Запрос успешно выполнен")
	}
}
