resource "signoz_planned_maintenance" "weekly_db" {
  name        = "weekly-db-maintenance"
  description = "Silence database alerts during the weekly maintenance window."
  alert_ids   = ["018f9b2a-1111-7000-8000-000000000001", "018f9b2a-1111-7000-8000-000000000002"]

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
