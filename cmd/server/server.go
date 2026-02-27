package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/audit"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/grpc"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/handler"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/delivery/http/router"
	memstorage "github.com/Nekrasov-Sergey/metrics-collector/internal/server/repository/mem_storage"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/repository/postgres"
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

	r, err := router.New(
		l,
		gin.ReleaseMode,
		router.WithSignKey(cfg.SignKey),
		router.WithCryptoKey(cfg.CryptoKey),
		router.WithTrustedSubnet(cfg.TrustedSubnet),
	)
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

	h := handler.New(s, auditLogger, l, handler.WithStoreInterval(cfg.StoreInterval))
	h.RegisterRoutes(r)

	if cfg.DatabaseDSN == "" {
		h.StartMetricSaver(ctx)
	}

	httpSrv := http.New(r, cfg.Addr, l)

	grpcSrv, err := grpc.New(s, auditLogger, l, grpc.WithGRPCAddress(cfg.GRPCAddr), grpc.WithTrustedSubnet(cfg.TrustedSubnet))
	if err != nil {
		return err
	}

	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.Run(); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcSrv.Run(); err != nil {
			errCh <- err
		}
	}()

	var runErr error

	select {
	case <-ctx.Done():
		l.Info().Msg("Получен сигнал завершения")
	case err := <-errCh:
		runErr = err
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	errChShutdown := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			errChShutdown <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcSrv.Shutdown(shutdownCtx); err != nil {
			errChShutdown <- err
		}
	}()

	wg.Wait()
	close(errChShutdown)

	for e := range errChShutdown {
		runErr = multierr.Append(runErr, e)
	}

	if cfg.DatabaseDSN == "" && cfg.StoreInterval > 0 {
		s.SaveMetricsToFile(shutdownCtx)
	}

	return runErr
}
