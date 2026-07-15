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

    selected_query_name = "A"
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
