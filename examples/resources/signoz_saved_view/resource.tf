# A saved view on the Logs explore page: errors for one service, rendered as a
# list with a handful of columns selected. See the schema below for all attributes.

resource "signoz_saved_view" "api_service_errors" {
  name           = "api-service-errors"
  source         = "logs"
  schema_version = "v2"

  spec = {
    display_name = "API service errors"
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

              filter = {
                expression = "service.name = 'api-service' AND severity_text = 'ERROR'"
              }

              order = [
                {
                  direction = "desc"
                  key = {
                    name = "timestamp"
                  }
                }
              ]

              limit = 100
            }
          }
        }
      }
    ]

    selected_fields = [
      {
        name = "timestamp"
      },
      {
        name            = "service.name"
        field_context   = "resource"
        field_data_type = "string"
      },
      {
        name            = "severity_text"
        field_data_type = "string"
      },
      {
        name            = "body"
        field_context   = "body"
        field_data_type = "string"
      },
    ]

    display = {
      font_size = "small"
      format    = "list"
      max_lines = 2
    }
  }
}
