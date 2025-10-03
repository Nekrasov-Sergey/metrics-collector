package agentconfig

import (
	"flag"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
)

type SecondDuration time.Duration

func (d *SecondDuration) String() string {
	return time.Duration(*d).String()
}

func (d *SecondDuration) Set(s string) error {
	seconds, err := strconv.Atoi(s)
	if err != nil {
		return errors.Wrap(err, "значение должно быть в секундах")
	}
	*d = SecondDuration(time.Duration(seconds) * time.Second)
	return nil
}

func (d *SecondDuration) UnmarshalText(text []byte) error {
	seconds, err := strconv.Atoi(string(text))
	if err != nil {
		return errors.Wrap(err, "значение должно быть в секундах")
	}
	*d = SecondDuration(time.Duration(seconds) * time.Second)
	return nil
}

type Config struct {
	Addr           string         `env:"ADDRESS"`
	PollInterval   SecondDuration `env:"POLL_INTERVAL"`
	ReportInterval SecondDuration `env:"REPORT_INTERVAL"`
}

func New() (*Config, error) {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	pollInterval := SecondDuration(2 * time.Second)
	reportInterval := SecondDuration(10 * time.Second)
	flag.Var(&pollInterval, "p", "частота опроса метрик из пакета runtime")
	flag.Var(&reportInterval, "r", "частота отправки метрик на сервер")

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
