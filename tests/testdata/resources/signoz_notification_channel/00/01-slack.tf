# Scenario 00 — slack notification channel, minimal config. Create, no-drift,
# destroy. Base-only scenarios may be authored in HCL (.tf); scenarios with JSON
# patches use a .tf.json base (see scenario 01 and ../../../README.md).
resource "signoz_notification_channel" "scenario_00" {
  name         = "slack.scenario-00"
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
