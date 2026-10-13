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

output "ingestion_key_value" {
  value     = data.signoz_ingestion_key.existing.value
  sensitive = true
}

output "ingestion_key_limits" {
  value = data.signoz_ingestion_key.existing.limits
}
