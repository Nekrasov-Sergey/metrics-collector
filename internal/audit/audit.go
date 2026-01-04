package audit

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

// Observer описывает получателя событий аудита.
//
// Реализации Observer отвечают за доставку событий во внешние системы (файл, HTTP и т.п.).
type Observer interface {
	// Notify обрабатывает событие аудита.
	Notify(ctx context.Context, event *types.AuditEvent) error
}

// Audit реализует диспетчеризацию событий аудита.
//
// Отправляет события всем зарегистрированным наблюдателям.
type Audit struct {
	observers []Observer
}

func New(cfg *config.ServerConfig) (*Audit, error) {
	observers := make([]Observer, 0, 2)

	if cfg.AuditFile != "" {
		fileObserver, err := NewFileObserver(cfg.AuditFile)
		if err != nil {
			return nil, err
		}
		observers = append(observers, fileObserver)
	}

	if cfg.AuditURL != "" {
		httpObserver := NewHTTPObserver(cfg.AuditURL)
		observers = append(observers, httpObserver)
	}

	return &Audit{observers: observers}, nil
}

// Info регистрирует событие аудита и передает его всем наблюдателям.
//
// Ошибки отдельных наблюдателей логируются и не прерывают обработку остальных.
func (s *Audit) Info(ctx context.Context, event *types.AuditEvent) {
	for _, observer := range s.observers {
		if err := observer.Notify(ctx, event); err != nil {
			log.Error().Err(err).Msg("Не удалось выполнить аудит события")
		}
	}
}
