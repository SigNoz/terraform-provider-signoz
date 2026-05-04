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

resource "signoz_route_policy" "warnings" {
  name        = "warnings-to-test-channel"
  description = "Route warning-severity alerts to the test channel."
  expression  = "severity == \"warning\""
  channels    = ["test-channel", "second-channel"]
  kind        = "policy"
  tags        = ["acc-test", "v2"]
}

output "route_policy_id" {
  value = signoz_route_policy.warnings.id
}
