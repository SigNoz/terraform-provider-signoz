resource "signoz_auth_domain" "saml" {
  name    = "corp.example.com"
  enabled = true
  config = {
    saml = {
      kind = "saml"
      spec = {
        entity_id   = "https://idp.example.com/metadata"
        location    = "https://idp.example.com/sso"
        certificate = file("${path.module}/idp-cert.pem")
      }
    }
  }
}
