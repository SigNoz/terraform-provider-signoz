# Rule created in the disabled state, using the `not_equal` operator and an
# explicitly disabled renotify block.
resource "signoz_rule" "disabled_renotify_off" {
  alert      = "testdata-disabled-renotify-off"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"
  disabled   = true

  condition = {
    composite_query = {
      panel_type = "value"
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
                    metric_name       = "up"
                    space_aggregation = "min"
                    time_aggregation  = "latest"
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
            name       = "target down"
            op         = "not_equal"
            target     = 1
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
      enabled = false
    }
  }

  schema_version = "v2alpha1"
}
