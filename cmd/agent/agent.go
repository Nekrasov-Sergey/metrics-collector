package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	agentconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	buildinfo "github.com/Nekrasov-Sergey/metrics-collector/pkg/build_info"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildinfo.Print(types.BuildInfo{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	})
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Агент завершился с ошибкой")
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	l := logger.New()
	client := resty.New()

	cfg, err := agentconfig.NewAgentConfig(l)
	if err != nil {
		return err
	}

	a, err := agent.New(cfg, client, l)
	if err != nil {
		return err
	}

	return a.Run(ctx)
}
