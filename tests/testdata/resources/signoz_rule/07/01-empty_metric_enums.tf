# The rule below references this channel by display name; defining it here
# replaces the suite's old hand-seeded channels.
resource "signoz_notification_channel" "slack" {
  display_name = "slack-scenario-07"

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
  alert      = "testdata-empty-metric-enums"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  condition = {
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
                source = ""

                aggregations = [
                  {
                    metric_name       = "system.memory.utilization"
                    space_aggregation = "avg"
                    time_aggregation  = ""
                    temporality       = ""
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
            channels   = [signoz_notification_channel.slack.display_name]
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
