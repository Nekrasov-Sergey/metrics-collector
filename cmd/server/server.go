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
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	buildinfo "github.com/Nekrasov-Sergey/metrics-collector/pkg/build_info"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildinfo.Print(types.BuildInfo{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	})
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Сервер завершился с ошибкой")
	}
}

func run() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
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

	r, err := router.New(l, gin.ReleaseMode, router.WithSignKey(cfg.SignKey), router.WithCryptoKey(cfg.CryptoKey))
	if err != nil {
		return err
	}

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

	s := service.New(
		ctx,
		repo,
		l,
		service.WithStoreInterval(cfg.StoreInterval),
		service.WithRestore(cfg.Restore),
		service.WithFileStoragePath(cfg.FileStoragePath),
	)
	h := rest.New(s, auditLogger, l, rest.WithStoreInterval(cfg.StoreInterval))
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

	if cfg.DatabaseDSN == "" && cfg.StoreInterval > 0 {
		s.SaveMetricsToFile(shutdownCtx)
	}

	l.Info().Msg("Сервер остановлен")
	return nil
}
