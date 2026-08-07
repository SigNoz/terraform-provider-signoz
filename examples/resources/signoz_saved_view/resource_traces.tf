# A saved view on the Traces explore page: slow checkout spans, grouped by
# operation and rendered as a table.

resource "signoz_saved_view" "checkout_slow_spans" {
  name   = "checkout-slow-spans"
  source = "traces"

  data = {
    schema_version = "v2"

    spec = {
      display_name = "Checkout slow spans"
      panel_type   = "table"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name   = "A"
                signal = "traces"

                aggregations = [
                  {
                    expression = "p99(duration_nano)"
                  }
                ]

                filter = {
                  expression = "service.name = 'checkout' AND duration_nano > 500000000"
                }

                group_by = [
                  {
                    name            = "name"
                    field_context   = "span"
                    field_data_type = "string"
                  }
                ]

                limit = 50
              }
            }
          }
        }
      ]

      selected_fields = [
        {
          name            = "name"
          field_context   = "span"
          field_data_type = "string"
        },
        {
          name            = "duration_nano"
          field_context   = "span"
          field_data_type = "float64"
        },
      ]

      display = {
        format = "table"
      }
    }
  }
}
