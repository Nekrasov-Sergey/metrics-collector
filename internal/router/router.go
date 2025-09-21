package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func New(l zerolog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), LoggerMiddleware(l))
	return r
}
