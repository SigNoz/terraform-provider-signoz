terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint = "http://localhost:3301"
  # access_token = "ACCESS_TOKEN"
}

data "signoz_dashboard" "get_dashboard" {
  uuid = "<uuid>"
}

output "dashboard" {
  value = data.signoz_dashboard.get_dashboard
}
