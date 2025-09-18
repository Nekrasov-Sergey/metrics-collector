package memstorage

import (
	"sync"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type MemStorage struct {
	sync.RWMutex
	Metrics map[types.MetricName]types.Metric
}

func New() *MemStorage {
	return &MemStorage{
		Metrics: make(map[types.MetricName]types.Metric),
	}
}
