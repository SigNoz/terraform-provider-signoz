# Email notification channel, a second config variant of the exactly-one union.
resource "signoz_notification_channel" "scenario_02" {
  name         = "email-scenario-02"
  display_name = "Scenario 02 Email"

  config = {
    email = {
      kind = "email"
      spec = {
        to = "oncall-scenario-02@example.com"
      }
    }
  }
}
