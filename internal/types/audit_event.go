package types

type AuditEvent struct {
	TS        int64        `json:"ts"`
	Metrics   []MetricName `json:"metrics"`
	IPAddress string       `json:"ip_address"`
}
