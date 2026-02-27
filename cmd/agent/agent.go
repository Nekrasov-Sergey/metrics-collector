package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent/grpc"
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

func run() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	l := logger.New()

	cfg, err := agentconfig.NewAgentConfig(l)
	if err != nil {
		return err
	}

	httpClient := resty.New()

	grpcClient, err := grpc.New(l, grpc.WithGRPCAddress(cfg.GRPCAddr), grpc.WithLocalIP(cfg.LocalIP))
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(grpcClient))

	a := agent.New(cfg, httpClient, grpcClient.Client, l)

	return a.Run(ctx)
}
