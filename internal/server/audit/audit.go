package audit

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type Observer interface {
	Notify(ctx context.Context, event *types.AuditEvent) error
}

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

func (s *Audit) Info(ctx context.Context, event *types.AuditEvent) {
	for _, observer := range s.observers {
		if err := observer.Notify(ctx, event); err != nil {
			log.Error().Err(err).Msg("Не удалось выполнить аудит события")
		}
	}
}
