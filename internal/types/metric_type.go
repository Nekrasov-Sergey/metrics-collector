package types

type MetricType string

func (t MetricType) IsValid() bool {
	return t == Gauge || t == Counter
}

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
