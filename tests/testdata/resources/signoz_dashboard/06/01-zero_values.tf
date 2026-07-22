resource "signoz_dashboard" "zero_values" {
  schema_version = "v6"
  name           = "testdata-dashboard-zero-values-qp84zc"
  tags = [
    {
      key   = "tag"
      value = "testdata"
    }
  ]
  spec = {
    display = {
      name        = "Zero and empty values"
      description = "Probes zero/empty values the API may drop, incl. the empty-string display description it marshals omitempty."
    }
    variables = []
    panels = {
      "22222222-2222-4222-8222-222222222222" = {
        kind = "Panel"
        spec = {
          display = {
            name        = "Empty and zero fields"
            description = ""
          }
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
                chart_appearance = {
                  line_interpolation = "spline"
                  show_points        = false
                  line_style         = "solid"
                  fill_mode          = "none"
                  span_gaps = {
                    fill_only_below = false
                    fill_less_than  = "0s"
                  }
                }
                axes = {
                  soft_min     = 0
                  soft_max     = 0
                  is_log_scale = false
                }
                legend = {
                  position = "bottom"
                  mode     = "list"
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
                    spec = {
                      metrics = {
                        name          = "A"
                        step_interval = "60"
                        signal        = "metrics"
                        aggregations = [
                          {
                            metric_name       = "system.memory.usage"
                            time_aggregation  = "avg"
                            space_aggregation = "sum"
                            reduce_to         = "avg"
                          }
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
                  ref = "#/spec/panels/22222222-2222-4222-8222-222222222222"
                }
              }
            ]
          }
        }
      }
    ]
  }
}
