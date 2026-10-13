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

resource "signoz_downtime_schedule" "weekly_maintenance" {
  name        = "Weekly maintenance"
  description = "Suppress alerts during weekly maintenance window"
  alert_ids   = ["5", "10"]
  all_alerts  = false

  schedule {
    start_time = "2024-01-15T02:00:00Z"
    end_time   = "2024-01-15T04:00:00Z"
    timezone   = "UTC"

    recurrence {
      type         = "weekly"
      days_of_week = [0, 6]
      duration     = "2h"
    }
  }
}

resource "signoz_downtime_schedule" "one_time" {
  name        = "One-time deployment window"
  description = "Suppress all alerts during deployment"
  all_alerts  = true

  schedule {
    start_time = "2024-02-01T00:00:00Z"
    end_time   = "2024-02-01T06:00:00Z"
    timezone   = "America/New_York"

    recurrence {
      type = "once"
    }
  }
}

output "downtime_schedule" {
  value = signoz_downtime_schedule.weekly_maintenance
}
