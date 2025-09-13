package logger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const loggerKey ctxKey = "logger"

// New настраивает формат вывода zerolog
func New() zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		FormatCaller: func(i interface{}) string {
			if s, ok := i.(string); ok {
				lastSlash := strings.LastIndex(s, "/")
				if lastSlash != -1 {
					s = s[lastSlash+1:]
				}
				return s
			}
			return ""
		},
		FormatExtra: func(m map[string]interface{}, b *bytes.Buffer) error {
			keys := []string{"method", "url", "host", "req_id", "req_time_ms", "status"}
			for _, key := range keys {
				if val, ok := m[key]; ok {
					if _, err := fmt.Fprintf(b, " \033[36m%s\033[0m=%v", key, val); err != nil {
						return err
					}
				}
			}
			return nil
		},
		FieldsExclude: []string{"method", "url", "host", "req_id", "req_time_ms", "status"},
	}

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger
	return logger
}

// GinMiddleware добавляет zerolog в gin
func GinMiddleware(baseLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startLogger := baseLogger.With().
			Str("method", c.Request.Method).
			Str("url", c.Request.URL.Path).
			Str("host", c.Request.Host).
			Str("req_id", uuid.NewString()[:8]).
			Logger()
		startLogger.Info().Msg("Start handling request ")

		// Кладём логгер в обычный контекст
		ctx := context.WithValue(c.Request.Context(), loggerKey, &startLogger)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		endLogger := startLogger.With().
			Dur("req_time_ms", time.Since(start)).
			Int("status", c.Writer.Status()).
			Logger()

		if len(c.Errors) > 0 {
			endLogger.Error().Err(c.Errors[0].Err).Msg("Failed handling request")
			return
		}
		endLogger.Info().Msg("Finish handling request")
	}
}

type errorOptions struct {
	enableJSON   bool
	enableLogger bool
}

type ErrorOpt func(*errorOptions)

func WithoutJSON() ErrorOpt {
	return func(o *errorOptions) {
		o.enableJSON = false
	}
}

func WithoutLogger() ErrorOpt {
	return func(o *errorOptions) {
		o.enableLogger = false
	}
}

// Error обрабатывает ошибку: логирует её через zerolog и/или возвращает JSON-ответ.
// Поведение можно изменить с помощью опций (например, WithoutJSON или WithoutLogger).
func Error(c *gin.Context, err error, status int, opts ...ErrorOpt) {
	options := &errorOptions{
		enableJSON:   true,
		enableLogger: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	if options.enableJSON {
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
	}

	if options.enableLogger {
		_ = c.Error(err)
	}
}

func C(ctx context.Context) *zerolog.Logger {
	logger, ok := ctx.Value(loggerKey).(*zerolog.Logger)
	if !ok {
		return &log.Logger
	}
	return logger
}
