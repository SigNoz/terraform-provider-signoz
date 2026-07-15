# Metric rule using the `below` operator with the `on_average` match type, a
# `table` panel, and absent-data alerting (alert_on_absent + absent_for).
resource "signoz_rule" "scenario_04" {
  alert      = "testdata-metrics-below-absent"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  condition = {
    alert_on_absent = true
    absent_for      = 5

    composite_query = {
      panel_type = "table"
      query_type = "builder"
      unit       = "percent"

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
                    metric_name       = "system.memory.utilization"
                    space_aggregation = "avg"
                    time_aggregation  = "avg"
                  }
                ]

                filter = {
                  expression = "state = 'free'"
                }
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
            match_type = "on_average"
            name       = "low free memory"
            op         = "below"
            target     = 10
          }
        ]
      }
    }
  }

  evaluation = {
    rolling = {
      kind = "rolling"
      spec = {
        eval_window = "10m"
        frequency   = "5m"
      }
    }
  }

  notification_settings = {}

  schema_version = "v2alpha1"
}
