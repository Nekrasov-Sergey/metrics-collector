package main

import (
	"os"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
)

func main() {
	if err := agent.Run(); err != nil {
		os.Exit(1)
	}
}
