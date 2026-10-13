package model

// DataRetention maps the GET response from /api/v1/settings/ttl.
// Uses the expected_* fields to avoid drift during TTL change processing.
type DataRetention struct {
	MetricsTTLDurationHrs     int `json:"expected_metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs int `json:"expected_metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        int `json:"expected_logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    int `json:"expected_logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      int `json:"expected_traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  int `json:"expected_traces_move_ttl_duration_hrs"`
}

// DataRetentionUpdate is the payload for setting retention.
// v2 API uses duration_hrs (integer hours).
type DataRetentionUpdate struct {
	Type                   string `json:"type"`
	DurationHrs            int64  `json:"duration_hrs"`
	ColdStorageDurationHrs int64  `json:"cold_storage_duration_hrs,omitempty"`
}
