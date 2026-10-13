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

data "signoz_notification_channel" "get_channel" {
  id = "1"
}

output "channel" {
  value = data.signoz_notification_channel.get_channel
}
