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

resource "signoz_ingestion_key" "production" {
  name       = "Production Ingest Key"
  tags       = ["production", "team-a"]
  expires_at = ""

  rate_limits = {
    day_limit    = 1000000
    second_limit = 1000
  }
}

output "ingestion_key" {
  value     = signoz_ingestion_key.production
  sensitive = true
}
