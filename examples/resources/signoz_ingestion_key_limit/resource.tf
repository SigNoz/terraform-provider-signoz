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
  name = "Production Ingest Key"
  tags = ["production", "team-a"]
}

# Bound production logs to 1M requests / 1 GB per day and 1k requests per second.
resource "signoz_ingestion_key_limit" "logs_daily" {
  key_id = signoz_ingestion_key.production.id
  signal = "logs"

  config = {
    day = {
      count = 1000000
      size  = 1073741824
    }
    second = {
      count = 1000
    }
  }

  tags = ["tier:gold"]
}
