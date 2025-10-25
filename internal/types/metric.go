package types

type Metric struct {
	Name  MetricName `json:"id" db:"name"`
	MType MetricType `json:"type" db:"type"`
	Delta *int64     `json:"delta,omitempty" db:"delta"`
	Value *float64   `json:"value,omitempty" db:"value"`
}
