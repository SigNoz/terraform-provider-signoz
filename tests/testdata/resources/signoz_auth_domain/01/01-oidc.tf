# OIDC auth domain (issuer is required for the oidc config variant).
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
