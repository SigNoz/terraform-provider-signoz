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

resource "signoz_log_pipeline" "extract_severity" {
  name        = "Extract Severity"
  alias       = "extract-severity"
  description = "Extract severity from log body"
  enabled     = true
  order_id    = 1

  filter = jsonencode({
    op = "AND"
    items = [
      {
        key = {
          key  = "service_name"
          type = "tag"
        }
        op    = "="
        value = "api-server"
      }
    ]
  })

  config = jsonencode([
    {
      id         = "proc-1"
      orderId    = 1
      enabled    = true
      name       = "Extract severity"
      type       = "grok_parser"
      output     = "severity"
      parse_from = "body"
      pattern    = "%{WORD:severity}"
    }
  ])
}

output "log_pipeline" {
  value = signoz_log_pipeline.extract_severity
}
