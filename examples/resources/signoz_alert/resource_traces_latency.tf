# Traces threshold — p99 latency with ns → s unit conversion.
# The series unit is ns (compositeQuery.unit) but the target is in seconds
# (threshold.targetUnit). SigNoz converts before comparing.
resource "signoz_alert" "traces_threshold_latency" {
  alert       = "Search API p99 latency above 5s"
  alert_type  = "TRACES_BASED_ALERT"
  description = "p99 duration of the search endpoint exceeds 5s"
  summary     = "Search-api latency degraded"
  severity    = "warning"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "ns"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "traces"
            stepInterval = 60
            aggregations = [{ expression = "p99(duration_nano)" }]
            filter = {
              expression = "service.name = 'search-api' AND name = 'GET /api/v1/search'"
            }
            groupBy = [
              { name = "service.name", fieldContext = "resource", fieldDataType = "string" },
              { name = "http.route", fieldContext = "attribute", fieldDataType = "string" },
            ]
            legend = "{{service.name}} {{http.route}}"
          }
        }
      ]
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name       = "warning"
          op         = "above"
          matchType  = "at_least_once"
          target     = 5
          targetUnit = "s"
          channels   = ["slack"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "5m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["service.name", "http.route"]
    renotify = {
      enabled      = true
      interval     = "30m"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "search"
  }
}
