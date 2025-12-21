package audit

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type HTTPObserver struct {
	url    string
	client *resty.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: resty.New(),
	}
}

func (o *HTTPObserver) Notify(ctx context.Context, event *types.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "не удалось спарсить событие для аудита")
	}

	resp, err := o.client.R().SetContext(ctx).SetBody(data).Post(o.url)
	if err != nil {
		return errors.Wrapf(err, "не удалось отправить событие в сервис аудита по адресу %s", o.url)
	}

	if resp.StatusCode() >= 400 {
		return errors.Errorf("неуспешный код ответа сервиса аудита: %d", resp.StatusCode())
	}

	return nil
}
