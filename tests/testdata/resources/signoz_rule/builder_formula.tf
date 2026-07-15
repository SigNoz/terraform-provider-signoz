# Rule alerting on a formula that combines two builder queries. The alert is
# evaluated on the formula (selected_query_name = "F"), not on A or B directly.
resource "signoz_rule" "builder_formula" {
  alert      = "testdata-builder-formula"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "percentunit"

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
                    metric_name      = "http.server.requests.errors"
                    time_aggregation = "rate"
                    temporality      = "delta"
                  }
                ]
              }
            }
          }
        },
        {
          builder_query = {
            type = "builder_query"
            spec = {
              metrics = {
                name   = "B"
                signal = "metrics"

                aggregations = [
                  {
                    metric_name      = "http.server.requests.total"
                    time_aggregation = "rate"
                    temporality      = "delta"
                  }
                ]
              }
            }
          }
        },
        {
          builder_formula = {
            type = "builder_formula"
            spec = {
              name       = "F"
              expression = "A / B"
              legend     = "error ratio"
            }
          }
        }
      ]
    }

    selected_query_name = "F"

    thresholds = {
      basic = {
        kind = "basic"
        spec = [
          {
            channels   = ["pagerduty"]
            match_type = "all_the_times"
            name       = "high error ratio"
            op         = "above"
            target     = 0.05
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

  notification_settings = {}

  schema_version = "v2alpha1"
}
