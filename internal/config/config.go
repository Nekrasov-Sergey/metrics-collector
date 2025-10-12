package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return errors.New("адрес должен быть в формате host:port")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "неверный порт")
	}
	host := parts[0]
	if host == "" {
		host = "localhost"
	}
	a.Host = host
	a.Port = port
	return nil
}

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
