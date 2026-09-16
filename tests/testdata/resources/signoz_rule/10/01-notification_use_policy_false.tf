# notification_settings.use_policy = false must round-trip, not come back null.
# The rule below references this channel by display name; defining it here
# replaces the suite's old hand-seeded channels.
resource "signoz_notification_channel" "slack" {
  display_name = "slack-scenario-10"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook"
        send_resolved = true
      }
    }
  }
}

resource "signoz_rule" "scenario_07" {
  alert          = "testdata-use-policy-false"
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
            channels   = [signoz_notification_channel.slack.display_name]
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
    use_policy = false
  }
}
