package rest

import (
	"context"
	"time"
)

func (h *Handler) StartMetricSaver(ctx context.Context) {
	if h.config.StoreInterval > 0 {
		go func() {
			storeTicker := time.NewTicker(time.Duration(h.config.StoreInterval))
			for {
				select {
				case <-ctx.Done():
					h.logger.Info().Msg("Сохранение метрик в файл остановлено")
					return
				case <-storeTicker.C:
					h.service.SaveMetricsToFile(ctx)
				}
			}
		}()
	}
}
