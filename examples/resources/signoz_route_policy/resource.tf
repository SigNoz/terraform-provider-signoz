# Route policies reference channels by display name.
resource "signoz_notification_channel" "oncall_slack" {
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

resource "signoz_route_policy" "critical_to_oncall" {
  name        = "route-critical-to-oncall"
  description = "Route critical payments alerts to the on-call channel."
  kind        = "policy"
  expression  = "service == \"payments\" && severity == \"critical\""
  channels    = [signoz_notification_channel.oncall_slack.display_name]
  tags        = ["team:platform"]
}
