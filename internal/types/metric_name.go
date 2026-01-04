package types

// MetricName представляет имя метрики сервиса.
type MetricName string

// Набор поддерживаемых имен метрик.
//
// Используется для идентификации метрик при сборе, хранении и передаче данных между компонентами сервиса.
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
	TotalMemory   MetricName = "TotalMemory"
	FreeMemory    MetricName = "FreeMemory"
)
