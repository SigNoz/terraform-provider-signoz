# Roundtrip probe: `notification_settings.renotify` set to an empty object.
#
# renotify is Optional + Computed. The provider sends `renotify: {}`; the
# backend's Renotify.Enabled is tagged `json:"enabled"` (no omitempty), so the
# response echoes `{enabled: false}` while interval/alertStates stay absent.
# This probes whether an all-null renotify object round-trips cleanly or drifts.
resource "signoz_rule" "scenario_09" {
  alert          = "testdata-renotify-empty"
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
    renotify = {}
  }
}
