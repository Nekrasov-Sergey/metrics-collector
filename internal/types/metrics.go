package types

import (
	"time"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
//type Metrics struct {
//	ID    string   `json:"id"`
//	MType string   `json:"type"`
//	Delta *int64   `json:"delta,omitempty"`
//	Value *float64 `json:"value,omitempty"`
//	Hash  string   `json:"hash,omitempty"`
//}

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MetricName string

const (
	Alloc         MetricName = "Alloc"
	BuckHashSys   MetricName = "BuckHashSys"
	Frees         MetricName = "Frees"
	GCCPUFraction MetricName = "GCCPUFraction"
	GCSys         MetricName = "GCSys"
	HeapAlloc     MetricName = "HeapAlloc"
	HeapIdle      MetricName = "HeapIdle"
	HeapInuse     MetricName = "HeapInuse"
	HeapObjects   MetricName = "HeapObjects"
	HeapReleased  MetricName = "HeapReleased"
	HeapSys       MetricName = "HeapSys"
	LastGC        MetricName = "LastGC"
	Lookups       MetricName = "Lookups"
	MCacheInuse   MetricName = "MCacheInuse"
	MCacheSys     MetricName = "MCacheSys"
	MSpanInuse    MetricName = "MSpanInuse"
	MSpanSys      MetricName = "MSpanSys"
	Mallocs       MetricName = "Mallocs"
	NextGC        MetricName = "NextGC"
	NumForcedGC   MetricName = "NumForcedGC"
	NumGC         MetricName = "NumGC"
	OtherSys      MetricName = "OtherSys"
	PauseTotalNs  MetricName = "PauseTotalNs"
	StackInuse    MetricName = "StackInuse"
	StackSys      MetricName = "StackSys"
	Sys           MetricName = "Sys"
	TotalAlloc    MetricName = "TotalAlloc"
	RandomValue   MetricName = "RandomValue"
	PollCount     MetricName = "PollCount"
)

type Metric struct {
	Name  MetricName
	Type  MetricType
	Value float64
}

const (
	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second
)
