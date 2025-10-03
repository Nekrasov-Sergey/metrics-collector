package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
)

func main() {
	if err := agent.Run(); err != nil {
		log.Logger.Err(err).Msg("агент завершился с ошибкой")
		os.Exit(1)
	}
}
