# Minimal rule accepted by the v2 rules API. For schema_version = "v2alpha1"
# the server requires a query, thresholds (each with at least one channel to
# route to), evaluation, and notification_settings. Optional fields (labels,
# annotations, description, ...) are omitted.
resource "signoz_rule" "minimal" {
  alert          = "minimal-required-only"
  alert_type     = "METRIC_BASED_ALERT"
  rule_type      = "promql_rule"
  schema_version = "v2alpha1"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"

      queries = [
        {
          promql = {
            type = "promql"
            spec = {
              name  = "A"
              query = "up"
            }
          }
        }
      ]
    }

    thresholds = {
      basic = {
        kind = "basic"
        spec = [
          {
            name       = "critical"
            op         = "above"
            match_type = "at_least_once"
            target     = 1
            channels   = ["slack"]
          }
        ]
      }
    }
  }

  evaluation = {
    rolling = {
      kind = "rolling"
      spec = {
        eval_window = "5m"
        frequency   = "1m"
      }
    }
  }

  notification_settings = {}
}
