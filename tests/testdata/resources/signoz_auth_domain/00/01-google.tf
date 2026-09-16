# Google auth domain with minimal required config.
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
