package main

import (
	"os"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}
