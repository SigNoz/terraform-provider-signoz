# Roundtrip probe: `notification_settings.renotify.alert_states` set to [].
#
# alert_states is Optional + Computed. The provider sends `alertStates: []`,
# but the backend field is tagged `json:"alertStates,omitempty"`, so an empty
# slice is dropped from the response and flattens back to null. Expected: the
# create fails with
# `.notification_settings.renotify.alert_states: was [], but now null` — the
# same drop-the-empty-slice class of bug that group_by (scenario 08) hit before
# GroupBy switched to `omitzero`.
resource "signoz_rule" "scenario_11" {
  alert          = "testdata-alert-states-empty"
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
    renotify = {
      alert_states = []
    }
  }
}
