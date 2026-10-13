# A logs-based threshold alert. The fork's signoz_alert keeps the v1 rules API:
# `condition` is the composite-query JSON document the server expects.

resource "signoz_alert" "api_error_rate" {
  alert      = "High API Error Rate"
  alert_type = "LOGS_BASED_ALERT"
  severity   = "critical"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      builderQueries = {
        A = {
          dataSource        = "logs"
          queryName         = "A"
          aggregateOperator = "count"
        }
      }
    }
    op        = "1"
    target    = 100
    matchType = "1"
  })
}
