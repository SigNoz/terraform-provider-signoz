# Rule backed by a raw ClickHouse SQL query instead of the query builder.
resource "signoz_rule" "scenario_03" {
  alert      = "testdata-clickhouse-sql"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  condition = {
    composite_query = {
      panel_type = "table"
      query_type = "clickhouse_sql"

      queries = [
        {
          clickhouse_sql = {
            type = "clickhouse_sql"
            spec = {
              name   = "A"
              legend = "rows"
              query  = "SELECT toStartOfMinute(now()) AS ts, count() AS value FROM signoz_logs.distributed_logs_v2 GROUP BY ts"
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
            name       = "critical"
            op         = "above"
            target     = 1000
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
