resource "signoz_planned_maintenance" "weekly_all" {
  name        = "weekly-maintenance"
  description = "Silence all alerts during the weekly maintenance window."

  schedule = {
    timezone   = "UTC"
    start_time = "2026-01-01T02:00:00Z"
    end_time   = "2026-01-01T04:00:00Z"

    recurrence = {
      repeat_type = "weekly"
      repeat_on   = ["sunday"]
      duration    = "2h0m0s"
    }
  }
}
