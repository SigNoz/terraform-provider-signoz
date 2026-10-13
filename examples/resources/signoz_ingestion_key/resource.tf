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
  expires_at = "2027-01-01T00:00:00Z"
}

resource "signoz_ingestion_key" "permanent" {
  name = "Permanent Ingest Key"
}

output "ingestion_key_value" {
  value     = signoz_ingestion_key.production.value
  sensitive = true
}
