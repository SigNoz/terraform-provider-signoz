# Smallest valid signoz_dashboard: a single number panel in a single grid layout,
# no variables. Exercises the required-attribute floor (schema_version, name,
# tags, spec.display, spec.variables, spec.panels, spec.layouts).
resource "signoz_dashboard" "minimal" {
  schema_version = "v6"
  name           = "testdata-dashboard-minimal-mn19qd"
  tags           = []

  spec = {
    display = {
      name = "Minimal dashboard"
    }
    links     = []
    variables = []
    panels = {
      "00000000-0000-4000-8000-000000000001" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Uptime"
          }
          links = []
          plugin = {
            number_panel = {
              kind = "signoz/NumberPanel"
              spec = {
                visualization = {
                  time_preference = "global_time"
                }
                formatting = {
                  unit              = "s"
                  decimal_precision = "2"
                }
              }
            }
          }
          queries = [
            {
              kind = "scalar"
              spec = {
                name = "A"
                plugin = {
                  builder_query = {
                    kind = "signoz/BuilderQuery"
                    spec = {
                      metrics = {
                        name          = "A"
                        step_interval = "60"
                        signal        = "metrics"
                        aggregations = [
                          {
                            metric_name       = "system.uptime"
                            time_aggregation  = "max"
                            space_aggregation = "max"
                            reduce_to         = "last"
                          },
                        ]
                        filter = {
                          expression = ""
                        }
                        having = {
                          expression = ""
                        }
                      }
                    }
                  }
                }
              }
            },
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
                width  = 6
                height = 4
                content = {
                  ref = "#/spec/panels/00000000-0000-4000-8000-000000000001"
                }
              },
            ]
          }
        }
      },
    ]
  }
}
