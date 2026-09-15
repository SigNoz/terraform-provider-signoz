# Scenario 02 — email notification channel. Proves a second `config` variant
# round-trips cleanly (the union is exactly-one of ten kinds).
resource "signoz_notification_channel" "scenario_02" {
  name         = "email.scenario-02"
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
