package config

import (
	"crypto/rsa"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	cryptokeys "github.com/Nekrasov-Sergey/metrics-collector/pkg/crypto_keys"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/network"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

// AgentConfig содержит конфигурацию агента сбора метрик.
type AgentConfig struct {
	HTTPAddr       string         `env:"ADDRESS" json:"address"`
	GRPCAddr       string         `env:"GRPC_ADDRESS" json:"grpc_address"`
	PollInterval   SecondDuration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval SecondDuration `env:"REPORT_INTERVAL" json:"report_interval"`
	Key            string         `env:"KEY" json:"key"`
	RateLimit      int            `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey      string         `env:"CRYPTO_KEY" json:"crypto_key"`
	PublicKey      *rsa.PublicKey
	LocalIP        string
}

func NewAgentConfig(logger zerolog.Logger) (*AgentConfig, error) {
	localIP, err := network.GetLocalIP()
	if err != nil {
		return nil, err
	}

	cfg := AgentConfig{
		HTTPAddr:       "localhost:8080",
		PollInterval:   SecondDuration(2 * time.Second),
		ReportInterval: SecondDuration(10 * time.Second),
		RateLimit:      1,
		LocalIP:        localIP,
	}

	var configPath string
	flag.StringVar(&configPath, "c", "", "имя файла конфигурации")
	flag.StringVar(&configPath, "config", "", "имя файла конфигурации")

	var addr NetAddress
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	grpcAddr := flag.String("g", "", "адрес GRPC-сервера")

	var pollInterval SecondDuration
	flag.Var(&pollInterval, "p", "частота опроса метрик из пакета runtime в секундах")

	var reportInterval SecondDuration
	flag.Var(&reportInterval, "r", "частота отправки метрик на сервер в секундах")

	key := flag.String("k", "", "ключ для вычисления хеша")
	rateLimit := flag.Int("l", 1, "количество одновременно исходящих запросов на сервер")
	cryptoKey := flag.String("crypto-key", "", "путь до файла с публичным ключом")

	flag.Parse()

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, errors.Wrap(err, "не удалось прочитать файл с конфигом")
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, errors.Wrap(err, "не удалось распарсить файл с конфигом")
		}
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.HTTPAddr = addr.String()
		case "g":
			cfg.GRPCAddr = utils.Deref(grpcAddr)
		case "p":
			cfg.PollInterval = pollInterval
		case "r":
			cfg.ReportInterval = reportInterval
		case "k":
			cfg.Key = utils.Deref(key)
		case "l":
			cfg.RateLimit = utils.Deref(rateLimit)
		case "crypto-key":
			cfg.CryptoKey = utils.Deref(cryptoKey)
		}
	})

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	publicKey, err := cryptokeys.GetPublicKey(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}
	cfg.PublicKey = publicKey

	logger.Info().
		Str("address", cfg.HTTPAddr).
		Str("grpc_address", cfg.GRPCAddr).
		Str("poll_interval", cfg.PollInterval.String()).
		Str("report_interval", cfg.ReportInterval.String()).
		Str("key", cfg.Key).
		Int("rate_limit", cfg.RateLimit).
		Str("crypto_key", cfg.CryptoKey).
		Msg("Загружена конфигурация агента")

	return &cfg, nil
}
