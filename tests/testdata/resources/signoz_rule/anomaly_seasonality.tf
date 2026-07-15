# Anomaly rule: alerts on deviation from a learned seasonal baseline rather than
# a fixed target. Exercises the condition fields specific to anomaly detection
# (algorithm, seasonality, require_min_points, required_num_points).
resource "signoz_rule" "anomaly_seasonality" {
  alert      = "testdata-anomaly-seasonality"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "anomaly_rule"

  condition = {
    algorithm           = "standard"
    seasonality         = "daily"
    require_min_points  = true
    required_num_points = 4

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
                    metric_name       = "http.server.request.count"
                    space_aggregation = "sum"
                    time_aggregation  = "rate"
                    temporality       = "delta"
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
            name       = "traffic anomaly"
            op         = "above"
            target     = 3
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
