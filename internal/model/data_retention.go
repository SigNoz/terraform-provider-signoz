package model

// DataRetention maps the merged GET responses for the retention settings:
// per-signal /api/v1/settings/ttl (expected_* fields avoid drift during TTL
// change processing) plus /api/v2/settings/ttl for logs. Hours.
type DataRetention struct {
	MetricsTTLDurationHrs     int `json:"expected_metrics_ttl_duration_hrs"`
	MetricsMoveTTLDurationHrs int `json:"expected_metrics_move_ttl_duration_hrs"`
	LogsTTLDurationHrs        int `json:"expected_logs_ttl_duration_hrs"`
	LogsMoveTTLDurationHrs    int `json:"expected_logs_move_ttl_duration_hrs"`
	TracesTTLDurationHrs      int `json:"expected_traces_ttl_duration_hrs"`
	TracesMoveTTLDurationHrs  int `json:"expected_traces_move_ttl_duration_hrs"`
}

// DataRetentionUpdate carries one signal's retention setting to the client.
// The client maps it onto the wire shapes: query params for
// /api/v1/settings/ttl (metrics/traces) and the days-based JSON body for
// /api/v2/settings/ttl (logs). Hours here; days on the wire.
type DataRetentionUpdate struct {
	Type               string
	DurationHrs        int64
	ColdStorageVolume  string
	ColdStorageMoveHrs int64
}

// CustomRetentionRule is a per-condition log retention rule; not managed by
// the provider (an empty list is sent and preserved server-side).
type CustomRetentionRule struct {
	ColumnName string `json:"columnName"`
	Operator   string `json:"operator"`
	Value      string `json:"value"`
}

// CustomRetentionTTLParams is the POST /api/v2/settings/ttl body (logs only).
type CustomRetentionTTLParams struct {
	Type                    string                `json:"type"`
	DefaultTTLDays          int64                 `json:"defaultTTLDays"`
	TTLConditions           []CustomRetentionRule `json:"ttlConditions"`
	ColdStorageVolume       string                `json:"coldStorageVolume,omitempty"`
	ColdStorageDurationDays int64                 `json:"coldStorageDurationDays,omitempty"`
}

// CustomRetentionTTLResponseData is the payload of GET /api/v2/settings/ttl.
type CustomRetentionTTLResponseData struct {
	Version                        string                `json:"version"`
	Status                         string                `json:"status"`
	ExpectedLogsTTLDurationHrs     int64                 `json:"expected_logs_ttl_duration_hrs"`
	ExpectedLogsMoveTTLDurationHrs int64                 `json:"expected_logs_move_ttl_duration_hrs"`
	DefaultTTLDays                 int64                 `json:"default_ttl_days,omitempty"`
	TTLConditions                  []CustomRetentionRule `json:"ttl_conditions,omitempty"`
	ColdStorageVolume              string                `json:"cold_storage_volume,omitempty"`
	ColdStorageTTLDays             int64                 `json:"cold_storage_ttl_days,omitempty"`
}
