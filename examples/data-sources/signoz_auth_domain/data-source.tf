data "signoz_auth_domain" "by_id" {
  id = signoz_auth_domain.google.id
}

output "auth_domain_sso_type" {
  value = data.signoz_auth_domain.by_id.sso_type
}
