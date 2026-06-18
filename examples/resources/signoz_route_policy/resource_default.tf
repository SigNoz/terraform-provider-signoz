resource "signoz_route_policy" "default_kind" {
  name        = "route-critical-to-oncall"
  expression  = "service == \"payments\" && severity == \"critical\""
  channels    = ["oncall-slack"]
}
