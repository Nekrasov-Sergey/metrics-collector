package server

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Server struct {
	httpServer *http.Server
	logger     zerolog.Logger
}

func New(handler http.Handler, addr string, logger zerolog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		logger: logger,
	}
}

func (s *Server) Run() error {
	s.logger.Info().Msgf("Сервер запущен на %s", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "ошибка при запуске сервера")
	}
	return nil
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return errors.Wrap(err, "ошибка при остановке сервера")
	}
	return nil
}
