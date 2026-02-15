package agent

import (
	"crypto/rsa"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	cryptokeys "github.com/Nekrasov-Sergey/metrics-collector/pkg/crypto_keys"
)

// Agent реализует агент сбора и отправки метрик.
//
// Агент периодически собирает системные метрики, агрегирует их и отправляет на сервер метрик.
type Agent struct {
	config    *config.AgentConfig
	client    *resty.Client
	logger    zerolog.Logger
	publicKey *rsa.PublicKey
}

func New(config *config.AgentConfig, client *resty.Client, logger zerolog.Logger) (*Agent, error) {
	agent := &Agent{
		config: config,
		client: client,
		logger: logger,
	}

	publicKey, err := cryptokeys.GetPublicKey(config.CryptoKey)
	if err != nil {
		return nil, err
	}
	agent.publicKey = publicKey

	return agent, nil
}
