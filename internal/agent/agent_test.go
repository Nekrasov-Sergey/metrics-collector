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

	restMocks "github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest/mocks"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	serviceMocks "github.com/Nekrasov-Sergey/metrics-collector/internal/server/service/mocks"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func TestRunAgent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	l := logger.New()

	type buildMock struct {
		repo  *serviceMocks.RepoMock
		audit *restMocks.AuditMock
	}
	tests := []struct {
		name  string
		build func(*buildMock)
	}{
		{
			name: "Success",
			build: func(m *buildMock) {
				m.repo.UpdateMetricsMock.Return(nil)
				m.repo.GetMetricsMock.Return(nil, nil)
				m.audit.InfoMock.Return()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			mock := &buildMock{
				repo:  serviceMocks.NewRepoMock(ctrl),
				audit: restMocks.NewAuditMock(ctrl),
			}
			tt.build(mock)

			r := router.New(l, gin.TestMode)
			s := service.New(ctx, mock.repo, l)
			h := rest.New(s, mock.audit, l)
			h.RegisterRoutes(r)
			srv := httptest.NewServer(r)
			defer srv.Close()

			agentCfg := &config.AgentConfig{
				Addr:           strings.TrimPrefix(srv.URL, "http://"),
				PollInterval:   config.SecondDuration(50 * time.Millisecond),
				ReportInterval: config.SecondDuration(100 * time.Millisecond),
				RateLimit:      5,
			}

			client := resty.New()

			a := agent.New(agentCfg, client, logger.New())

			runCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
			defer cancel()

			err := a.Run(runCtx)
			require.NoError(t, err)
		})
	}
}
