package types

// MetricType представляет тип метрики.
type MetricType string

func (t MetricType) IsValid() bool {
	return t == Gauge || t == Counter
}

// Поддерживаемые типы метрик.
const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
