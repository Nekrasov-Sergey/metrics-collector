package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/audit"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/rest"
	memstorage "github.com/Nekrasov-Sergey/metrics-collector/internal/server/repository/mem_storage"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/repository/postgres"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Сервер завершился с ошибкой")
	}
}

func run() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.New()

	cfg, err := config.NewServerConfig(l)
	if err != nil {
		return err
	}

	auditLogger, err := audit.New(cfg)
	if err != nil {
		return err
	}

	r := router.New(l, gin.ReleaseMode, cfg.Key)

	var repo service.Repository
	if cfg.DatabaseDSN != "" {
		repo, err = postgres.New(cfg.DatabaseDSN, l)
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&err, multierr.Close(repo))
	} else {
		repo = memstorage.New()
	}

	s := service.New(ctx, repo, cfg, l)
	h := rest.New(s, cfg, l, auditLogger)
	h.RegisterRoutes(r)

	if cfg.DatabaseDSN == "" {
		h.StartMetricSaver(ctx)
	}

	srv := server.New(r, cfg.Addr, l)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Run()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	l.Info().Msg("Сервер остановлен")
	return nil
}
