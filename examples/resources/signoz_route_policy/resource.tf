terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:3301"
  access_token = "<SIGNOZ-API-KEY>"
}

resource "signoz_route_policy" "critical_alerts" {
  route_name  = "critical-alerts"
  description = "Route critical alerts to PagerDuty"
  enabled     = true

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }

  channel_ids = ["1", "3"]

  continue_matching = false
  group_wait        = "30s"
  group_interval    = "5m"
  repeat_interval   = "4h"
}

output "route_policy" {
  value = signoz_route_policy.critical_alerts
}
