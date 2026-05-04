terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:8080"
  access_token = "<SIGNOZ-API-KEY>"
}

# A weekly recurring maintenance window: 2 hours every Mon/Wed/Fri,
# starting 1 June 2026, in UTC.
resource "signoz_planned_maintenance" "weekly" {
  name        = "weekly-deploy-window"
  description = "Weekly recurring downtime window."

  schedule = {
    timezone   = "UTC"
    start_time = "2026-06-01T00:00:00Z"

    recurrence = {
      duration    = "2h"
      repeat_type = "weekly"
      start_time  = "2026-06-01T00:00:00Z"
      repeat_on   = ["monday", "wednesday", "friday"]
    }
  }
}

output "planned_maintenance_id" {
  value = signoz_planned_maintenance.weekly.id
}
