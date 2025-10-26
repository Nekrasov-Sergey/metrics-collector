package agent_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/agent_config"
	serverconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service/mocks"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func TestRunAgent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	l := logger.New()

	r := router.New(l, gin.TestMode)

	srvConfig, err := serverconfig.New(l)
	require.NoError(t, err)

	ctrl := minimock.NewController(t)
	repo := mocks.NewRepoMock(ctrl)
	repo.UpdateMetricsMock.Return(nil)

	s := service.New(ctx, srvConfig, repo, l)
	h := rest.New(srvConfig, s, l)
	h.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()

	agentCfg := &agentconfig.Config{
		Addr:           strings.TrimPrefix(srv.URL, "http://"),
		PollInterval:   config.SecondDuration(50 * time.Millisecond),
		ReportInterval: config.SecondDuration(100 * time.Millisecond),
	}

	client := resty.New()

	a := agent.New(agentCfg, client, logger.New())

	err = a.Run(ctx)
	require.NoError(t, err)
}
