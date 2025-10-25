package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	agentconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Агент завершился с ошибкой")
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.New()
	client := resty.New()

	cfg, err := agentconfig.New()
	if err != nil {
		return err
	}

	a := agent.New(cfg, client, l)
	return a.Run(ctx)
}
