resource "signoz_saved_view" "minimal" {
  name           = "minimal-required-only"
  source         = "logs"
  schema_version = "v2"

  spec = {
    display_name = "Minimal"
    panel_type   = "list"
    request_type = "raw"

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
  }
}
