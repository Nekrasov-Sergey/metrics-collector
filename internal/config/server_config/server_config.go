package serverconfig

import (
	"flag"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
)

type Config struct {
	Addr string
}

func New() *Config {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	flag.Parse()

	return &Config{
		Addr: addr.String(),
	}
}
