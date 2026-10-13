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

data "signoz_log_pipeline" "get_pipeline" {
  id = "pipeline-abc-123"
}

output "log_pipeline" {
  value = data.signoz_log_pipeline.get_pipeline
}
