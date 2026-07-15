# Rule evaluated on a cumulative (scheduled) window rather than a rolling one —
# resets every week on the configured weekday/hour/minute in the given timezone.
resource "signoz_rule" "cumulative_evaluation" {
  alert      = "testdata-cumulative-weekly"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

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
                    metric_name      = "billing.usage.units"
                    time_aggregation = "increase"
                    temporality      = "cumulative"
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
            match_type = "in_total"
            name       = "weekly quota exceeded"
            op         = "above"
            target     = 100000
          }
        ]
      }
    }
  }

  evaluation = {
    cumulative = {
      kind = "cumulative"
      spec = {
        frequency = "1h"
        timezone  = "UTC"
        schedule = {
          type    = "weekly"
          weekday = 1
          hour    = 0
          minute  = 0
        }
      }
    }
  }

  notification_settings = {
    new_group_eval_delay = "1m"
  }

  schema_version = "v2alpha1"
}
