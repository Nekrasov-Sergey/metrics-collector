package agentconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
)

type Config struct {
	Addr           string                `env:"ADDRESS"`
	PollInterval   config.SecondDuration `env:"POLL_INTERVAL"`
	ReportInterval config.SecondDuration `env:"REPORT_INTERVAL"`
}

func New() (*Config, error) {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	pollInterval := config.SecondDuration(2 * time.Second)
	reportInterval := config.SecondDuration(10 * time.Second)
	flag.Var(&pollInterval, "p", "частота опроса метрик из пакета runtime в секундах")
	flag.Var(&reportInterval, "r", "частота отправки метрик на сервер в секундах")

	flag.Parse()

	cfg := Config{
		Addr:           addr.String(),
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	return &cfg, nil
}
