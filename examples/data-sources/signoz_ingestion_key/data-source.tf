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

data "signoz_ingestion_key" "existing" {
  id = "key-abc-123"
}

output "ingestion_key" {
  value     = data.signoz_ingestion_key.existing
  sensitive = true
}
