package server

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
)

type Server struct {
	httpServer *http.Server
	logger     zerolog.Logger
}

func New(handler http.Handler, config *serverconfig.Config, logger zerolog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    config.Addr,
			Handler: handler,
		},
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		s.logger.Info().Msgf("Сервер запущен на %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("Ошибка при запуске сервера")
			cancel()
		}
	}()

	<-ctx.Done()
	s.logger.Info().Msg("Остановка сервера...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		return err
	}

	s.logger.Info().Msg("Сервер остановлен")
	return nil
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return errors.Wrap(err, "ошибка при остановке сервера")
	}
	return nil
}
