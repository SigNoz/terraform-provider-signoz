package attr

const (
	MetricsTTLDuration     = "metrics_ttl_duration_hrs"
	MetricsMoveTTLDuration = "metrics_move_ttl_duration_hrs"
	LogsTTLDuration        = "logs_ttl_duration_hrs"
	LogsMoveTTLDuration    = "logs_move_ttl_duration_hrs"
	TracesTTLDuration      = "traces_ttl_duration_hrs"
	TracesMoveTTLDuration  = "traces_move_ttl_duration_hrs"

	Metrics                = "metrics"
	Logs                   = "logs"
	Traces                 = "traces"
	TTLDurationHrs         = "ttl_duration_hrs"
	ColdStorageDurationHrs = "cold_storage_duration_hrs"
	ColdStorageVolume      = "cold_storage_volume"
)
