terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:8080"
  access_token = "<SIGNOZ-API-KEY>"
}

# Google OAuth — exactly one of google_auth_config / oidc_config /
# saml_config must be set inside `config`.
resource "signoz_auth_domain" "google" {
  name = "example.com"
  config = {
    sso_type = "google_auth"
    google_auth_config = {
      client_id     = "9999999999-example.apps.googleusercontent.com"
      client_secret = "GOOGLE-CLIENT-SECRET"
    }
  }
}

# OIDC — same shape, different nested config.
resource "signoz_auth_domain" "oidc" {
  name = "example.org"
  config = {
    sso_type = "oidc"
    oidc_config = {
      issuer        = "https://idp.example.org"
      client_id     = "signoz"
      client_secret = "OIDC-CLIENT-SECRET"
    }
  }
}

output "auth_domain_id" {
  value = signoz_auth_domain.google.id
}
