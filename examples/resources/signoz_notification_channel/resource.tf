resource "signoz_notification_channel" "slack_oncall" {
  name         = "oncall-slack"
  display_name = "On-call Slack"

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
