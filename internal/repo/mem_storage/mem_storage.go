package memstorage

import (
	"sync"
)

type MemStorage struct {
	sync.RWMutex
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func New() *MemStorage {
	return &MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}
