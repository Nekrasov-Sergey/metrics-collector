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
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func TestRunAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := httptest.NewServer(gin.New())
	defer ts.Close()

	cfg := &agentconfig.Config{
		Addr:           ts.URL,
		PollInterval:   config.SecondDuration(50 * time.Millisecond),
		ReportInterval: config.SecondDuration(100 * time.Millisecond),
	}

	client := resty.New()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	a := agent.New(cfg, client, logger.New())
	err := a.Run(ctx)
	require.NoError(t, err)
}
