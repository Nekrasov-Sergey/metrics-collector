package memstorage

import (
	"context"
	"sync"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

// MemStorage реализует in-memory хранилище метрик.
//
// Используется для хранения метрик в оперативной памяти с защитой от конкурентного доступа.
type MemStorage struct {
	mu      *sync.Mutex
	metrics map[types.MetricName]*types.Metric
}

func New() *MemStorage {
	return &MemStorage{
		mu:      &sync.Mutex{},
		metrics: make(map[types.MetricName]*types.Metric),
	}
}

func (m *MemStorage) Ping(_ context.Context) error {
	return nil
}

func (m *MemStorage) Close() error {
	return nil
}
