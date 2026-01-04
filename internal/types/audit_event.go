package types

// AuditEvent описывает событие аудита сервиса.
//
// Содержит информацию о времени события, затронутых метриках и IP-адресе источника запроса.
type AuditEvent struct {
	TS        int64        `json:"ts"`
	Metrics   []MetricName `json:"metrics"`
	IPAddress string       `json:"ip_address"`
}
