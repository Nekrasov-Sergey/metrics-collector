package memstorage

import (
	"sync"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type MemStorage struct {
	sync.RWMutex
	metrics map[types.MetricName]types.Metric
}

func New() *MemStorage {
	return &MemStorage{
		metrics: make(map[types.MetricName]types.Metric),
	}
}
