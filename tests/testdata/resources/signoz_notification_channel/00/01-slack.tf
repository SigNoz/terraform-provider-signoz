# Slack notification channel with minimal config.
resource "signoz_notification_channel" "scenario_00" {
  name         = "slack-scenario-00"
  display_name = "Scenario 00 Slack"

  config = {
    slack = {
      kind = "slack"
      spec = {
        api_url = "https://example.com/slack-webhook"
        channel = "#alerts-scenario-00"
      }
    }
  }
}
