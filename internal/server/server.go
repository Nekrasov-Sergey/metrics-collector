package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/handler"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/repo/mem_storage"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/service"
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
	srv := service.New(memStorage)
	h := handler.New(srv)
	h.RegisterRoutes(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	l.Info().Msg("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		l.Fatal().Err(err).Msg("Error starting server")
	}
}
