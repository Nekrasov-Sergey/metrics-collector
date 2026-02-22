package agent

import (
	"crypto/rsa"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	cryptokeys "github.com/Nekrasov-Sergey/metrics-collector/pkg/crypto_keys"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/network"
)

// Agent реализует агент сбора и отправки метрик.
//
// Агент периодически собирает системные метрики, агрегирует их и отправляет на сервер метрик.
type Agent struct {
	config    *config.AgentConfig
	client    *resty.Client
	logger    zerolog.Logger
	publicKey *rsa.PublicKey
	localIP   string
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

	localIP, err := network.GetLocalIP()
	if err != nil {
		return nil, err
	}
	agent.localIP = localIP

	return agent, nil
}
