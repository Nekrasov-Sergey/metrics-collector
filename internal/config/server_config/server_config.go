package serverconfig

import (
	"flag"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
)

type Config struct {
	Addr string `env:"ADDRESS"`
}

func New() (*Config, error) {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	flag.Parse()

	cfg := Config{
		Addr: addr.String(),
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	return &cfg, nil
}
