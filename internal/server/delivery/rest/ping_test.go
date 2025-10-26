package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	serverconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service/mocks"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func TestHandler_ping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	l := logger.New()

	cfg := &serverconfig.Config{}

	type args struct {
		url string
	}
	type buildMock struct {
		repo *mocks.RepoMock
	}
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		args  args
		build func(*buildMock)
		want  want
	}{
		{
			name: "SuccessPing",
			args: args{
				url: "/ping",
			},
			build: func(m *buildMock) {
				m.repo.PingMock.Return(nil)
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "ErrorPing",
			args: args{
				url: "/ping",
			},
			build: func(m *buildMock) {
				m.repo.PingMock.Return(errors.New("БД недоступна"))
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			mock := &buildMock{
				repo: mocks.NewRepoMock(ctrl),
			}
			tt.build(mock)

			r := router.New(l, gin.TestMode)

			s := service.New(ctx, cfg, mock.repo, l)
			h := rest.New(cfg, s, l)
			h.RegisterRoutes(r)

			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()
			resp, err := client.R().Get(srv.URL + tt.args.url)
			require.NoError(t, err)
			require.Equal(t, tt.want.code, resp.StatusCode())
			if tt.want.body != "" {
				require.Contains(t, resp.String(), tt.want.body)
			}
		})
	}
}
