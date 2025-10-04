package logger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CtxKey string

const LogKey CtxKey = "logger"

// New настраивает формат вывода zerolog
func New() zerolog.Logger {
	keys := []string{"method", "url", "req_id", "status", "duration", "size", "stack"}
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		FormatExtra: func(m map[string]interface{}, b *bytes.Buffer) error {
			for _, key := range keys {
				if val, ok := m[key]; ok {
					if _, err := fmt.Fprintf(b, " \033[36m%s\033[0m=%v", key, val); err != nil {
						return err
					}
				}
			}
			return nil
		},
		FieldsExclude: keys,
	}

	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return fmt.Sprintf("%+v", err)
	}

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger()

	log.Logger = logger
	return logger
}

// Error обрабатывает ошибку: логирует её через zerolog и формирует JSON-ответ
func Error(c *gin.Context, err error, status int) {
	c.JSON(status, gin.H{
		"error": err.Error(),
	})
	_ = c.Error(err)
}

func C(ctx context.Context) *zerolog.Logger {
	logger, ok := ctx.Value(LogKey).(*zerolog.Logger)
	if !ok {
		return &log.Logger
	}
	return logger
}
