# Every op and match_type name, which the server must store and read back as written (#161).
resource "signoz_notification_channel" "slack" {
  name         = "slack-scenario-13"
  display_name = "slack-scenario-13"

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

resource "signoz_rule" "scenario_13" {
  alert          = "testdata-threshold-enum-names"
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
            name       = "above"
            op         = "above"
            target     = 1
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "all_the_times"
            name       = "below"
            op         = "below"
            target     = 2
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "on_average"
            name       = "equal"
            op         = "equal"
            target     = 3
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "in_total"
            name       = "not_equal"
            op         = "not_equal"
            target     = 4
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "last"
            name       = "above_or_equal"
            op         = "above_or_equal"
            target     = 5
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "at_least_once"
            name       = "below_or_equal"
            op         = "below_or_equal"
            target     = 6
          },
          {
            channels   = [signoz_notification_channel.slack.display_name]
            match_type = "all_the_times"
            name       = "outside_bounds"
            op         = "outside_bounds"
            target     = 7
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
}
