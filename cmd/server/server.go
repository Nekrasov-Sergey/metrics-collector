package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	serverconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
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

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.New()
	r := router.New(l)

	cfg, err := serverconfig.New(l)
	if err != nil {
		return err
	}

	var repo service.Repository
	if cfg.DatabaseDSN != "" {
		repo, err = postgres.New(cfg.DatabaseDSN, l)
		if err != nil {
			return err
		}
		defer repo.CloseDB()
	} else {
		repo = memstorage.New()
	}

	s := service.New(ctx, cfg, repo, l)
	h := rest.New(cfg, s, l)
	h.RegisterRoutes(r)

	if cfg.DatabaseDSN == "" {
		h.StartMetricSaver(ctx)
	}

	srv := server.New(r, cfg.Addr, l)
	return srv.Run(ctx)
}
