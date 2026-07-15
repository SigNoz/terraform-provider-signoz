# Logs-based rule: a builder query over the logs signal, counting error logs
# grouped by service, on a `value` panel.
resource "signoz_rule" "logs_count" {
  alert      = "testdata-logs-error-count"
  alert_type = "LOGS_BASED_ALERT"
  rule_type  = "threshold_rule"

  labels = {
    severity = "warning"
  }

  condition = {
    composite_query = {
      panel_type = "value"
      query_type = "builder"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              logs = {
                name   = "A"
                signal = "logs"

                aggregations = [
                  {
                    alias      = "error_count"
                    expression = "count()"
                  }
                ]

                filter = {
                  expression = "severity_text = 'ERROR'"
                }

                group_by = [
                  {
                    field_context   = "resource"
                    field_data_type = "string"
                    name            = "service.name"
                  }
                ]

                legend        = "{{service.name}}"
                step_interval = "60"
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
            name       = "too many errors"
            op         = "above"
            target     = 100
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
    group_by = ["service.name"]
  }

  schema_version = "v2alpha1"
}
