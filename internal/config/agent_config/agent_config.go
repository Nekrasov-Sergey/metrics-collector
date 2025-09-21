package agentconfig

import (
	"flag"
	"net/url"
	"strconv"
	"time"

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

type Config struct {
	Addr           string
	PollInterval   SecondDuration
	ReportInterval SecondDuration
}

func New() *Config {
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

	u := url.URL{
		Scheme: "http",
		Host:   addr.String(),
	}

	return &Config{
		Addr:           u.String(),
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
	}
}
