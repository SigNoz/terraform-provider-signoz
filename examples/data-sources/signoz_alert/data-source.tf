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

# Alerts are fetched by ID (from the alert's URL in the Signoz UI);
# everything else is computed.
data "signoz_alert" "high_latency" {
  id = "0192a9c2-1234-7abc-8def-0123456789ab"
}
