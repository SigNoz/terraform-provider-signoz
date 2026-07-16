# Regression: `notification_settings.use_policy` explicitly set to `false`.
#
# use_policy is Optional + Computed. When the user pins it to `false`, the
# provider must return `false` after apply — not `null`. The API drops the
# falsy `usePolicy` from its response, so a naive round-trip flattens the
# absent field back to null and Terraform rejects the apply with:
#
#   .notification_settings.use_policy: was cty.False, but now null
#
# This base-only scenario fails at `terraform apply` until the round-trip
# preserves the explicit `false`.
resource "signoz_rule" "scenario_07" {
  alert          = "testdata-use-policy-false"
  alert_type     = "METRIC_BASED_ALERT"
  rule_type      = "threshold_rule"
  schema_version = "v2alpha1"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              metrics = {
                name   = "A"
                signal = "metrics"

                aggregations = [
                  {
                    metric_name       = "system.cpu.utilization"
                    space_aggregation = "avg"
                    time_aggregation  = "avg"
                  }
                ]
              }
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
            channels   = ["slack"]
            match_type = "at_least_once"
            name       = "warning"
            op         = "above"
            target     = 0.7
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
    use_policy = false
  }
}
