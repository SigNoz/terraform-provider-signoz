# Roundtrip probe: `notification_settings.new_group_eval_delay` set to "".
#
# new_group_eval_delay is Optional + Computed and maps to a backend
# valuer.TextDuration parsed with Go's time.ParseDuration. An empty string is
# not a valid duration, so the create is expected to fail with a 400 from the
# API rather than a drift/inconsistent-result — i.e. "" never round-trips at
# all. A meaningful value ("2m") is exercised by the modified existing
# scenarios.
resource "signoz_rule" "scenario_10" {
  alert          = "testdata-new-group-eval-delay-empty"
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
    new_group_eval_delay = ""
  }
}
