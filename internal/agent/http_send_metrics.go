package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/goccy/go-json"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/cryptoutil"
)

func (a *Agent) httpSendMetrics(ctx context.Context, metrics []types.Metric, w int) {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось спарсить метрики в json")
		return
	}

	compressedMetrics, err := a.compressMetrics(metricsJSON)
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось сжать метрики")
		return
	}

	encryptedMetrics, err := a.encryptMetrics(compressedMetrics)
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось зашифровать метрики")
		return
	}

	path, err := url.JoinPath("http://", a.config.HTTPAddr, "/updates")
	if err != nil {
		a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось сформировать url")
		return
	}

	req := a.httpClient.R().
		SetContext(ctx).
		SetBody(encryptedMetrics).
		SetHeader("X-Real-IP", a.config.LocalIP).
		SetHeader("Content-Encoding", "gzip")

	if a.config.Key != "" {
		hashSHA256 := cryptoutil.HMACSHA256([]byte(a.config.Key), metricsJSON)
		req.SetHeader("HashSHA256", hashSHA256)
	}

	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}
	for i, delay := range delays {
		resp, err := req.Post(path)
		if err == nil {
			if resp.IsError() {
				a.logger.Error().Int("воркер", w).Msgf("Сервер вернул ошибку: %s", resp.String())
				return
			}
			a.logger.Info().Int("воркер", w).Msgf("Отправлены метрики на http-сервер %s", a.config.HTTPAddr)
			return
		}

		var urlErr *url.Error
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			a.logger.Error().Err(err).Int("воркер", w).Msg("Истек таймаут запроса")
			continue
		case errors.Is(err, context.Canceled):
			a.logger.Error().Err(err).Int("воркер", w).Msg("Контекст отменен")
			continue
		case errors.As(err, &urlErr):
			a.logger.Error().Err(err).Int("воркер", w).Msgf("Сервер недоступен, попытка №%d", i+1)
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					a.logger.Error().Int("воркер", w).Msg("Запрос отменён контекстом во время ожидания")
					continue
				case <-timer.C:
				}
			}
		default:
			a.logger.Error().Err(err).Int("воркер", w).Msg("Неизвестная ошибка")
			continue
		}
	}

	a.logger.Error().Int("воркер", w).Msg("Все попытки отправки исчерпаны")
}

func (a *Agent) compressMetrics(metricsJSON []byte) ([]byte, error) {
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)

	if _, err := zw.Write(metricsJSON); err != nil {
		return nil, errors.Wrap(err, "не удалось записать данные для сжатия")
	}

	if err := zw.Close(); err != nil {
		return nil, errors.Wrap(err, "не удалось сжать данные")
	}

	return b.Bytes(), nil
}

func (a *Agent) encryptMetrics(metrics []byte) ([]byte, error) {
	if a.config.PublicKey == nil {
		return metrics, nil
	}

	encryptedMetrics, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, a.config.PublicKey, metrics, nil)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось зашифровать метрики rsa ключом")
	}

	return encryptedMetrics, nil
}
