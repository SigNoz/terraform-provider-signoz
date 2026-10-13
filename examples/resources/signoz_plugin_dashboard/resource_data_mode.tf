# Lacrosse devops board: pinned plugin version, element auto-filled from the
# catalog manifest, stub data passed via properties.
resource "signoz_plugin_dashboard" "devops" {
  title       = "Fleet DevOps"
  description = "Terraform-managed devops WebComponent dashboard."
  tags        = ["lacrosse", "devops", "plugins"]

  page {
    name = "Fleet"

    plugin_panel {
      title          = "Fleet node grid"
      plugin_id      = "fleet-node-grid"
      plugin_version = "0.2.0"
      x              = 0
      y              = 0
      w              = 8
      h              = 6
      properties = jsonencode({
        mode = "stub"
        stub = {
          nodes = ["edge-01", "edge-02"]
        }
      })
    }

    plugin_panel {
      title     = "System status"
      plugin_id = "system-status-board"
      x         = 8
      y         = 0
      w         = 4
      h         = 6
      attributes = {
        "data-source" = "stub"
      }
    }
  }
}
