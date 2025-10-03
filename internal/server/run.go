package server

import (
	serverconfig "github.com/Nekrasov-Sergey/metrics-collector/internal/config/server_config"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/handler"
	memstorage "github.com/Nekrasov-Sergey/metrics-collector/internal/repository/mem_storage"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/router"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/service"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

func Run() error {
	l := logger.New()
	r := router.New(l)

	config, err := serverconfig.New()
	if err != nil {
		return err
	}

	memStorage := memstorage.New()
	srv := service.New(memStorage)
	h := handler.New(srv)
	h.RegisterRoutes(r)

	server := New(r, config, l)
	return server.Run()
}
