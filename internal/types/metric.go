package types

type Metric struct {
	Name  MetricName `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
	//Hash  string   `json:"hash,omitempty"`
}
