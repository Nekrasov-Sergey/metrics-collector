package router

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type options struct {
	cryptoKey string
	signKey   string
}

type Option func(*options)

func WithSignKey(signKey string) Option {
	return func(o *options) {
		o.signKey = signKey
	}
}

func WithCryptoKey(cryptoKey string) Option {
	return func(o *options) {
		o.cryptoKey = cryptoKey
	}
}

func New(logger zerolog.Logger, mode string, opts ...Option) *gin.Engine {
	gin.SetMode(mode)

	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	r := gin.New()
	r.Use(gin.Recovery(), LoggerMiddleware(logger), DecryptMiddleware(o.cryptoKey), CompressMiddleware(), SignatureMiddleware(o.signKey))
	pprof.Register(r)
	return r
}
