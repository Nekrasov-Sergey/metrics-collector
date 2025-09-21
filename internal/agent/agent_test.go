package agent_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	agentconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func TestRunAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := httptest.NewServer(gin.New())
	defer ts.Close()

	config := &agentconfig.Config{
		Addr:           ts.URL,
		PollInterval:   agentconfig.SecondDuration(100 * time.Millisecond),
		ReportInterval: agentconfig.SecondDuration(200 * time.Millisecond),
	}

	client := resty.New()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	a := agent.New(client, config, logger.New())
	err := a.Run(ctx)
	require.NoError(t, err)
}
