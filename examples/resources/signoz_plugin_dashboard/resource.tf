# Manage a SigNoz v2 dashboard of WebComponent plugin panels.
# devops-status ships in the SigNoz plugin catalog (plugins/devops-status);
# check your server's catalog via GET /api/v1/plugins for available ids.
resource "signoz_plugin_dashboard" "devops_status" {
  title       = "DevOps Status"
  description = "Real-time infrastructure and application status dashboard"
  tags        = ["devops", "team:platform"]

  page {
    name = "Overview"

    plugin_panel {
      title     = "System Status"
      plugin_id = "devops-status"
      element   = "lax-devops-status"

      x = 0
      y = 0
      w = 6
      h = 4

      # HTML attributes on the custom element (the plugin reads env)
      attributes = {
        env = "prod"
      }
    }

    plugin_panel {
      title     = "Alerts"
      plugin_id = "devops-status"
      element   = "lax-devops-status"

      x = 6
      y = 0
      w = 6
      h = 4
    }
  }
}
