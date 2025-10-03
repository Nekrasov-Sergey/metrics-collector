package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Logger.Err(err).Msg("сервер завершился с ошибкой")
		os.Exit(1)
	}
}
