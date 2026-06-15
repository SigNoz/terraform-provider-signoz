resource "signoz_route_policy" "critical_to_oncall" {
  name        = "route-critical-to-oncall"
  description = "Route critical payments alerts to the on-call channel."
  kind        = "policy"
  expression  = "service == \"payments\" && severity == \"critical\""
  channels    = ["oncall-slack"]
  tags        = ["team:platform"]
}
