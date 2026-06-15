# Logs threshold with builder formula (error rate percentage).
# Two disabled log-count queries (A = errors, B = total) combined via a
# builder_formula into a percentage — the classic service-level error-rate shape.
resource "signoz_alert" "logs_error_rate_formula" {
  alert       = "Payments-api error log rate above 1%"
  alert_type  = "LOGS_BASED_ALERT"
  description = "Error log ratio as a percentage of total logs for payments-api"
  summary     = "Payments-api error rate above {{$threshold}}%"
  severity    = "critical"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "percent"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "logs"
            stepInterval = 60
            disabled     = true
            aggregations = [{ expression = "count()" }]
            filter = {
              expression = "service.name = 'payments-api' AND severity_text IN ['ERROR', 'FATAL']"
            }
            groupBy = [
              { name = "deployment.environment", fieldContext = "resource", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_query"
          spec = {
            name         = "B"
            signal       = "logs"
            stepInterval = 60
            disabled     = true
            aggregations = [{ expression = "count()" }]
            filter       = { expression = "service.name = 'payments-api'" }
            groupBy = [
              { name = "deployment.environment", fieldContext = "resource", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_formula"
          spec = {
            name       = "F1"
            expression = "(A / B) * 100"
            legend     = "{{deployment.environment}}"
          }
        }
      ]
    }
    selectedQueryName = "F1"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name      = "critical"
          op        = "above"
          matchType = "at_least_once"
          target    = 1
          channels  = ["slack"]
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
    group_by = ["deployment.environment"]
    renotify = {
      enabled      = true
      interval     = "30m"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "payments"
  }
}
