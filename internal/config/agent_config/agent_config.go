package agentconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

type Config struct {
	Addr           string                `env:"ADDRESS"`
	PollInterval   config.SecondDuration `env:"POLL_INTERVAL"`
	ReportInterval config.SecondDuration `env:"REPORT_INTERVAL"`
	Key            string                `env:"KEY"`
	RateLimit      int                   `env:"RATE_LIMIT"`
}

func New(logger zerolog.Logger) (*Config, error) {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	pollInterval := config.SecondDuration(2 * time.Second)
	reportInterval := config.SecondDuration(10 * time.Second)
	flag.Var(&pollInterval, "p", "частота опроса метрик из пакета runtime в секундах")
	flag.Var(&reportInterval, "r", "частота отправки метрик на сервер в секундах")

	key := flag.String("k", "", "ключ для вычисления хеша")

	rateLimit := flag.Int("l", 1, "количество одновременно исходящих запросов на сервер")

	flag.Parse()

	cfg := Config{
		Addr:           addr.String(),
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Key:            utils.Deref(key),
		RateLimit:      utils.Deref(rateLimit),
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	logger.Info().
		Str("address", cfg.Addr).
		Str("poll_interval", cfg.PollInterval.String()).
		Str("report_interval", cfg.ReportInterval.String()).
		Str("key", cfg.Key).
		Int("rate_limit", cfg.RateLimit).
		Msg("Загружена конфигурация агента")

	return &cfg, nil
}
