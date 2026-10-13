# Read current SigNoz data retention settings.
data "signoz_data_retention" "current" {}

output "metrics_retention_days" {
  value = data.signoz_data_retention.current.metrics_ttl_duration_hrs / 24
}
