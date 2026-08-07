resource "signoz_saved_view" "minimal" {
  name   = "minimal-required-only"
  source = "logs"

  data = {
    schema_version = "v2"

    spec = {
      display_name = "Minimal"
      panel_type   = "list"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              logs = {
                name   = "A"
                signal = "logs"
              }
            }
          }
        }
      ]

      selected_fields = []

      display = {}
    }
  }
}
