# Scenario 00 — google auth domain, minimal required config. Create, no-drift,
# destroy. Base-only scenarios may be authored in HCL (.tf); scenarios with JSON
# patches use a .tf.json base (see scenario 03 and ../../../README.md).
resource "signoz_auth_domain" "scenario_00" {
  name = "google.scenario-00.example.com"
  config = {
    google = {
      kind = "google"
      spec = {
        client_id     = "912345678901-abcdefghijklmnop.apps.googleusercontent.com"
        client_secret = "GOCSPX-scenario-00-secret"
      }
    }
  }
}
