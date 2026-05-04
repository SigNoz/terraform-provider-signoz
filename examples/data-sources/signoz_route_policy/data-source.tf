data "signoz_route_policy" "by_id" {
  id = signoz_route_policy.warnings.id
}

output "route_policy_kind" {
  value = data.signoz_route_policy.by_id.kind
}

output "route_policy_created_at" {
  value = data.signoz_route_policy.by_id.created_at
}
