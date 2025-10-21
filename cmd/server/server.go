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
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/server/service"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Err(err).Msg("Сервер завершился с ошибкой")
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l := logger.New()
	r := router.New(l)

	cfg, err := serverconfig.New()
	if err != nil {
		return err
	}

	s := service.New(ctx, cfg, memstorage.New())
	h := rest.New(s, cfg)
	h.RegisterRoutes(r)
	h.StartMetricSaver(ctx)

	srv := server.New(r, cfg, l)
	return srv.Run(ctx)
}
