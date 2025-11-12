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
)

func TestHandler_updateMetric(t *testing.T) {
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
			name: "SuccessGauge",
			args: args{
				url: "/update/",
				body: `
{
  "id": "LastGC",
  "type": "gauge",
  "value": 134.24
}`,
			},
			build: func(m *buildMock) {
				m.repo.UpdateMetricMock.Return(nil)
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "SuccessCounter",
			args: args{
				url: "/update/",
				body: `
{
  "id": "PollCount",
  "type": "counter",
  "delta": 500
}`,
			},
			build: func(m *buildMock) {
				m.repo.UpdateMetricMock.Return(nil)
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "NameNotFound",
			args: args{
				url: "/update/",
				body: `
{
  "type": "gauge",
  "value": 134.24
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
				url: "/update/",
				body: `
{
  "id": "LastGC",
  "type": "GAUGE",
  "value": 134.24
}`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "некорректный тип метрики: GAUGE",
			},
		},
		{
			name: "IncorrectGaugeValue",
			args: args{
				url: "/update/",
				body: `
{
  "id": "LastGC",
  "type": "gauge"
}`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "значение метрики Gauge не задано",
			},
		},
		{
			name: "IncorrectCounterValue",
			args: args{
				url: "/update/",
				body: `
{
  "id": "PollCount",
  "type": "counter"
}`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "значение метрики Counter не задано",
			},
		},
		{
			name: "InternalServerError",
			args: args{
				url: "/update/",
				body: `
{
  "id": "LastGC",
  "type": "gauge",
  "value": 134.24
}`,
			},
			build: func(m *buildMock) {
				m.repo.UpdateMetricMock.Return(errors.New("не удалось обновить метрику"))
			},
			want: want{
				code: http.StatusInternalServerError,
				body: http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "IncorrectBody",
			args: args{
				url:  "/update/",
				body: `[]`,
			},
			build: func(m *buildMock) {
			},
			want: want{
				code: http.StatusBadRequest,
				body: "не удалось распарсить тело запроса",
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
