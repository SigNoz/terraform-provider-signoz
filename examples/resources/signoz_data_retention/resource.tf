# Manage SigNoz data retention settings.
# This is a singleton resource — only one instance should exist.
resource "signoz_data_retention" "this" {
  metrics_ttl_duration_hrs      = 720  # 30 days
  metrics_move_ttl_duration_hrs = 0    # cold storage disabled
  logs_ttl_duration_hrs         = 360  # 15 days
  logs_move_ttl_duration_hrs    = 0
  traces_ttl_duration_hrs       = 168  # 7 days
  traces_move_ttl_duration_hrs  = 0
}
