package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/handler"
	memstorage "github.com/Nekrasov-Sergey/metrics-collector/internal/repo/mem_storage"
)

func TestHandler_UpdateMetric(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	type args struct {
		url string
	}
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "SuccessGauge",
			args: args{
				url: "/update/gauge/Alloc/12.3",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "SuccessCounter",
			args: args{
				url: "/update/counter/PollCount/5",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "NameNotFound",
			args: args{
				url: "/update/gauge//12.3",
			},
			want: want{
				code: http.StatusNotFound,
				body: "отсутствует имя метрики",
			},
		},
		{
			name: "IncorrectType",
			args: args{
				url: "/update/GAUGE/Alloc/12.3",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "некорректный тип метрики: GAUGE",
			},
		},
		{
			name: "IncorrectGaugeValue",
			args: args{
				url: "/update/gauge/Alloc/twelve",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "значение метрики не float64",
			},
		},
		{
			name: "IncorrectCounterValue",
			args: args{
				url: "/update/counter/PollCount/10.5",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "значение метрики не int64",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := gin.New()
			memStorage := memstorage.New()
			h := handler.New(memStorage)
			h.RegisterRoutes(r)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()
			resp, err := client.R().Post(srv.URL + tt.args.url)
			require.NoError(t, err)
			require.Equal(t, tt.want.code, resp.StatusCode())
			if tt.want.body != "" {
				require.Contains(t, resp.String(), tt.want.body)
			}
		})
	}
}
