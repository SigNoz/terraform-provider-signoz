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

data "signoz_saved_view" "get_view" {
  uuid = "abc-123-def"
}

output "saved_view" {
  value = data.signoz_saved_view.get_view
}
