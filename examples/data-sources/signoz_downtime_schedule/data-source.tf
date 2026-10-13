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

data "signoz_downtime_schedule" "get_schedule" {
  id = "1"
}

output "downtime_schedule" {
  value = data.signoz_downtime_schedule.get_schedule
}
