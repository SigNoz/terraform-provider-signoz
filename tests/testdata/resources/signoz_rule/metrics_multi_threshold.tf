# Metric threshold rule with two severity bands (critical + warning) routed to
# different channels, each with a recovery_target and target_unit.
resource "signoz_rule" "metrics_multi_threshold" {
  alert      = "testdata-metrics-multi-threshold"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    description = "Memory working set at {{$value}} on {{$k8s.pod.name}}"
    summary     = "Pod memory high"
  }

  labels = {
    severity = "critical"
    team     = "platform"
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "bytes"

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
                    metric_name       = "k8s.pod.memory.working_set"
                    space_aggregation = "max"
                    time_aggregation  = "avg"
                    temporality       = "unspecified"
                  }
                ]

                filter = {
                  expression = "k8s.namespace.name = 'prod'"
                }

                group_by = [
                  {
                    field_context   = "resource"
                    field_data_type = "string"
                    name            = "k8s.pod.name"
                  }
                ]

                legend        = "{{k8s.pod.name}}"
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
            channels        = ["pagerduty"]
            match_type      = "all_the_times"
            name            = "critical"
            op              = "above"
            target          = 2000000000
            recovery_target = 1800000000
            target_unit     = "bytes"
          },
          {
            channels        = ["slack"]
            match_type      = "at_least_once"
            name            = "warning"
            op              = "above"
            target          = 1000000000
            recovery_target = 900000000
            target_unit     = "bytes"
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
    group_by = ["k8s.pod.name"]
    renotify = {
      alert_states = ["firing", "nodata"]
      enabled      = true
      interval     = "30m"
    }
  }

  schema_version = "v2alpha1"
}
