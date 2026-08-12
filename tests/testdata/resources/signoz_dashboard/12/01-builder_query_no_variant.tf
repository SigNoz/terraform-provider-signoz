# Negative control for #154: builder_query.spec is a flattened oneOf over
# logs/metrics/traces, and none is set. The union must be rejected at plan time —
# left to reach the API it serialises as "spec": null, which the server reports
# as `invalid signal ""`, the error in the issue.
resource "signoz_dashboard" "builder_query_no_variant" {
  schema_version = "v6"
  name           = "testdata-dashboard-builder-query-no-variant-hj73dm"
  tags           = []

  spec = {
    display = {
      name = "Builder query with no variant"
    }
    links     = []
    variables = []
    panels = {
      "00000000-0000-4000-8000-000000000121" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Under-specified query"
          }
          links = []
          plugin = {
            time_series_panel = {
              kind = "signoz/TimeSeriesPanel"
              spec = {
                visualization = {
                  time_preference = "global_time"
                  fill_spans      = false
                }
                formatting = {
                  unit              = "none"
                  decimal_precision = "2"
                }
              }
            }
          }
          queries = [
            {
              kind = "time_series"
              spec = {
                name = "A"
                plugin = {
                  builder_query = {
                    kind = "signoz/BuilderQuery"
                    spec = {}
                  }
                }
              }
            }
          ]
        }
      }
    }
    layouts = [
      {
        grid = {
          kind = "Grid"
          spec = {
            items = [
              {
                x      = 0
                y      = 0
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/00000000-0000-4000-8000-000000000121"
                }
              }
            ]
          }
        }
      }
    ]
  }
}
