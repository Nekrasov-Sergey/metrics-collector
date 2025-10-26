package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func New(logger zerolog.Logger, mode string) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()
	r.Use(gin.Recovery(), LoggerMiddleware(logger), CompressMiddleware())
	return r
}
