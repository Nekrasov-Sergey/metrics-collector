package rest

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func (h *Handler) StartMetricSaver(ctx context.Context) {
	if h.config.StoreInterval > 0 {
		go func() {
			storeTicker := time.NewTicker(time.Duration(h.config.StoreInterval))
			for {
				select {
				case <-ctx.Done():
					log.Info().Msg("Сохранение метрик в файл остановлено")
					return
				case <-storeTicker.C:
					h.service.SaveMetricsToFile(ctx)
				}
			}
		}()
	}
}
