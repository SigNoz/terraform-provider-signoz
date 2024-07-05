terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}
provider "signoz" {}

data "signoz_alert" "example" {}
