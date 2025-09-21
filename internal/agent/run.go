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

	client := resty.New()
	config := agentconfig.New()
	l := logger.New()

	agent := New(client, config, l)
	return agent.Run(ctx)
}
