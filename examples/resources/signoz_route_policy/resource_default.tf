resource "signoz_notification_channel" "oncall_slack" {
  name         = "oncall-slack"
  display_name = "oncall-slack"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-oncall"
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
