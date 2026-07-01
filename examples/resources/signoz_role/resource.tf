resource "signoz_role" "editor" {
  name        = "editor"
  description = "Can view dashboards across the org"

  transaction_groups = [
    {
      relation = "read"

      object_group = {
        resource = {
          type = "telemetryresource"
          kind = "dashboard"
        }

        selectors = ["*"]
      }
    },
  ]
}
