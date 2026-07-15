# PromQL rule exercising the full promql spec (query, legend, step, stats), the
# `equal` operator with the `last` match type, and policy-routed notifications.
resource "signoz_rule" "promql_full" {
  alert      = "testdata-promql-full"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "promql_rule"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"

      queries = [
        {
          promql = {
            type = "promql"
            spec = {
              name   = "A"
              query  = "sum(rate(http_requests_total{code=~\"5..\"}[5m]))"
              legend = "5xx rate"
              step   = "60"
              stats  = true
            }
          }
        }
      ]
    }

    selected_query_name = "A"

    thresholds = {
      basic = {
        kind = "basic"
        spec = [
          {
            channels   = ["pagerduty"]
            match_type = "last"
            name       = "no 5xx expected"
            op         = "equal"
            target     = 0
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

  notification_settings = {
    group_by             = ["code"]
    new_group_eval_delay = "2m"
    use_policy           = true
    renotify = {
      alert_states = ["firing", "nodata"]
      enabled      = true
      interval     = "15m"
    }
  }

  schema_version = "v2alpha1"
}
