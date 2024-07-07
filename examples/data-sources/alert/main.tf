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

data "signoz_alert" "get_alert" {
  id = "5"
}

output "alert" {
  value = data.signoz_alert.get_alert
}
