package agent

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"

	agentconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.New()
	client := resty.New()

	config, err := agentconfig.New()
	if err != nil {
		return err
	}

	agent := New(client, config, l)
	return agent.Run(ctx)
}
