# Scenario 01 — oidc auth domain. Exercises the oidc config variant (issuer is
# required for oidc). Create, no-drift, destroy.
resource "signoz_auth_domain" "scenario_01" {
  name = "oidc.scenario-01.example.com"
  config = {
    oidc = {
      kind = "oidc"
      spec = {
        client_id     = "scenario-01-client"
        client_secret = "scenario-01-secret"
        issuer        = "https://idp.example.com"
      }
    }
  }
}
