package main

import (
	"context"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/agent"
)

func main() {
	agent.Run(context.Background())
}
