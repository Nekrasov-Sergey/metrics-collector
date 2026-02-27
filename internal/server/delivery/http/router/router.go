package router

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	cryptokeys "github.com/Nekrasov-Sergey/metrics-collector/pkg/crypto_keys"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/network"
)

type options struct {
	cryptoKey     string
	signKey       string
	trustedSubnet string
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

func WithTrustedSubnet(trustedSubnet string) Option {
	return func(o *options) {
		o.trustedSubnet = trustedSubnet
	}
}

func New(logger zerolog.Logger, mode string, opts ...Option) (*gin.Engine, error) {
	gin.SetMode(mode)

	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	privateKey, err := cryptokeys.GetPrivateKey(o.cryptoKey)
	if err != nil {
		return nil, err
	}

	trustedNet, err := network.ParseCIDR(o.trustedSubnet)
	if err != nil {
		return nil, err
	}

	r := gin.New()
	r.Use(
		gin.Recovery(),
		LoggerMiddleware(logger),
		CheckIPMiddleware(trustedNet),
		DecryptMiddleware(privateKey),
		CompressMiddleware(),
		SignatureMiddleware(o.signKey),
	)
	pprof.Register(r)
	return r, nil
}
