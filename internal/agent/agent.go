package agent

import (
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	pb "github.com/Nekrasov-Sergey/metrics-collector/internal/proto"
)

// Agent реализует агент сбора и отправки метрик.
//
// Агент периодически собирает системные метрики, агрегирует их и отправляет на сервер метрик.
type Agent struct {
	config     *config.AgentConfig
	httpClient *resty.Client
	grpcClient pb.MetricsClient
	logger     zerolog.Logger
}

func New(
	config *config.AgentConfig,
	httpClient *resty.Client,
	grpcClient pb.MetricsClient,
	logger zerolog.Logger,
) *Agent {
	return &Agent{
		config:     config,
		httpClient: httpClient,
		grpcClient: grpcClient,
		logger:     logger,
	}
}
