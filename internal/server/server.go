package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/handler"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/repo/mem_storage"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func NewGinRouter(l zerolog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), logger.GinMiddleware(l))
	return r
}

func Run() {
	l := logger.New()
	r := NewGinRouter(l)

	memStorage := memstorage.New()
	h := handler.New(memStorage)
	h.RegisterRoutes(r)

	l.Info().Msg("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		l.Error().Err(err).Msg("Error starting server")
	}
}
