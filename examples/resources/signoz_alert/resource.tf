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

resource "signoz_alert" "high_latency" {
  alert      = "High API latency"
  alert_type = "METRIC_BASED_ALERT"
  condition = jsonencode({
    op = "AND"
    items = [
      {
        key         = "signoz_latency_p99"
        op          = ">"
        value       = "1000"
        type        = "metric"
        unit        = "ms"
        aggregation = "p99"
      }
    ]
  })
  severity = "warning"
}
