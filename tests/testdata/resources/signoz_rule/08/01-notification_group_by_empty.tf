# Roundtrip probe: `notification_settings.group_by` set to an empty list.
#
# group_by is Optional + Computed. The provider sends `groupBy: []`, but the
# backend field is tagged `omitempty`, so an empty slice is dropped from the
# response and flattens back to null. Expected: the create fails with
# `.notification_settings.group_by: was [], but now null` — the same
# drop-the-falsy-value class of bug as use_policy (scenario 07).
resource "signoz_rule" "scenario_08" {
  alert          = "testdata-group-by-empty"
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
    group_by = []
  }
}
