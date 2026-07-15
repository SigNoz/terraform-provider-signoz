# Traces-based rule: a builder query over the traces signal with a p95 duration
# aggregation and a span-attribute filter.
resource "signoz_rule" "traces_latency" {
  alert      = "testdata-traces-p95-latency"
  alert_type = "TRACES_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    summary = "Checkout p95 latency above 750ms"
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "ns"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name   = "A"
                signal = "traces"

                aggregations = [
                  {
                    alias      = "p95_latency"
                    expression = "p95(duration_nano)"
                  }
                ]

                filter = {
                  expression = "service.name = 'checkout' AND kind_string = 'Server'"
                }

                group_by = [
                  {
                    field_context   = "resource"
                    field_data_type = "string"
                    name            = "service.name"
                  }
                ]

                legend = "p95"
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
            channels    = ["slack", "pagerduty"]
            match_type  = "at_least_once"
            name        = "slow checkout"
            op          = "above"
            target      = 750000000
            target_unit = "ns"
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
    renotify = {
      alert_states = ["firing"]
      enabled      = true
      interval     = "1h"
    }
  }

  schema_version = "v2alpha1"
}
