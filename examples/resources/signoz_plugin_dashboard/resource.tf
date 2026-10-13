# Manage a SigNoz v2 dashboard of WebComponent plugin panels.
resource "signoz_plugin_dashboard" "hello" {
  title = "Plugin Demo"

  page {
    name = "Overview"

    plugin_panel {
      title     = "Hello"
      plugin_id = "hello"
      element   = "lax-hello-wc"
    }
  }
}
