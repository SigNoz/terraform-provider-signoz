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

resource "signoz_alert" "existing_alert" { }

output "existing_alert" {
  value = signoz_alert.existing_alert
}
