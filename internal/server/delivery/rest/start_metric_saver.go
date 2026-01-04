package rest

import (
	"context"
	"time"
)

// StartMetricSaver запускает фоновое периодическое сохранение метрик в файл.
//
// Сохранение выполняется с интервалом, заданным в конфигурации StoreInterval.
// Если StoreInterval меньше либо равен нулю, фоновая задача не запускается.
func (h *Handler) StartMetricSaver(ctx context.Context) {
	if h.storeInterval > 0 {
		go func() {
			storeTicker := time.NewTicker(time.Duration(h.storeInterval))
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
