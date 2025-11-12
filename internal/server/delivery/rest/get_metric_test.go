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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	serverconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service/mocks"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
)

func TestHandler_getMetric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	l := zerolog.Logger{}

	cfg := &serverconfig.Config{
		StoreInterval: 1,
	}

	type args struct {
		url  string
		body string
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
			name: "Success",
			args: args{
				url: "/value/",
				body: `
{
  "id": "LastGC",
  "type": "gauge"
}`,
			},
			build: func(m *buildMock) {
				m.repo.GetMetricMock.Return(types.Metric{}, nil)
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "NameNotFound",
			args: args{
				url: "/value/",
				body: `
{
  "type": "gauge"
}`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusNotFound,
				body: "отсутствует имя метрики",
			},
		},
		{
			name: "IncorrectType",
			args: args{
				url: "/value/",
				body: `
{
  "id": "LastGC",
  "type": "GAUGE"
}`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "некорректный тип метрики",
			},
		},
		{
			name: "IncorrectBody",
			args: args{
				url:  "/value/",
				body: `[]`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "не удалось распарсить тело запроса",
			},
		},
		{
			name: "MetricNotFound",
			args: args{
				url: "/value/",
				body: `
{
  "id": "LastGC",
  "type": "gauge"
}`,
			},
			build: func(m *buildMock) {
				m.repo.GetMetricMock.Return(types.Metric{}, errcodes.ErrMetricNotFound)
			},
			want: want{
				code: http.StatusNotFound,
				body: errcodes.ErrMetricNotFound.Error(),
			},
		},
		{
			name: "InternalServerError",
			args: args{
				url: "/value/",
				body: `
{
  "id": "LastGC",
  "type": "gauge"
}`,
			},
			build: func(m *buildMock) {
				m.repo.GetMetricMock.Return(types.Metric{}, errors.New("не удалось получить метрику"))
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

			r := router.New(l, gin.TestMode, "")

			s := service.New(ctx, cfg, mock.repo, l)
			h := rest.New(cfg, s, l)
			h.RegisterRoutes(r)

			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()
			resp, err := client.R().SetBody(tt.args.body).Post(srv.URL + tt.args.url)
			require.NoError(t, err)
			require.Equal(t, tt.want.code, resp.StatusCode())
			if tt.want.body != "" {
				require.Contains(t, resp.String(), tt.want.body)
			}
		})
	}
}
