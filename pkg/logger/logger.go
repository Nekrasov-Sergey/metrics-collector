package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

// RespondError обрабатывает ошибку и формирует JSON-ответ
func RespondError(c *gin.Context, err error, status int) {
	if err == nil {
		c.AbortWithStatus(status)
		return
	}

	if status >= 500 {
		c.AbortWithStatusJSON(status, gin.H{
			"error": http.StatusText(status),
		})
	} else {
		c.AbortWithStatusJSON(status, gin.H{
			"error": err.Error(),
		})
	}

	_ = c.Error(err)
}
