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

data "signoz_route_policy" "get_route_policy" {
  id = "1"
}

output "route_policy" {
  value = data.signoz_route_policy.get_route_policy
}
