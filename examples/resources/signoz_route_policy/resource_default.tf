resource "signoz_notification_channel" "oncall_slack" {
  name         = "oncall-slack"
  display_name = "oncall-slack"

  config = {
    slack = {
      kind = "slack"
      spec = {
        api_url       = "https://example.com/slack-webhook"
        channel       = "#alerts-oncall"
        send_resolved = true
      }
    }
  }
}

resource "signoz_route_policy" "default_kind" {
  name       = "route-critical-to-oncall"
  expression = "service == \"payments\" && severity == \"critical\""
  channels   = [signoz_notification_channel.oncall_slack.display_name]
}
