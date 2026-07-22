resource "signoz_dashboard" "omitempty_fields" {
  schema_version = "v6"
  name           = "testdata-dashboard-omitempty-fields-nf3wql"
  tags = [
    {
      key   = "tag"
      value = "testdata"
    }
  ]
  spec = {
    display = {
      name        = "Omitempty and nullable fields"
      description = "Round-trips fields the v2 API marshals omitempty/omitzero, populated so they must survive: dashboard/panel links, text-variable constant, list-variable custom_all_value/capturing_regexp/sort."
    }
    links = [
      {
        name             = "SigNoz"
        url              = "https://signoz.io"
        target_blank     = true
        render_variables = false
      }
    ]
    variables = [
      {
        text_variable = {
          kind = "TextVariable"
          spec = {
            name     = "constant_text"
            value    = "prod"
            constant = true
            display = {
              name        = "Environment"
              description = "Constant text variable"
            }
          }
        }
      },
      {
        list_variable = {
          kind = "ListVariable"
          spec = {
            name             = "service_list"
            allow_all_value  = true
            allow_multiple   = true
            custom_all_value = "__all__"
            capturing_regexp = "(.*)"
            sort             = "alphabetical-asc"
            display = {
              name        = "Service"
              description = "List variable with custom all value"
            }
            plugin = {
              custom_variable = {
                kind = "signoz/CustomVariable"
                spec = {
                  custom_value = "frontend,backend"
                }
              }
            }
          }
        }
      }
    ]
    panels = {
      "33333333-3333-4333-8333-333333333333" = {
        kind = "Panel"
        spec = {
          display = {
            name        = "Omitempty probe panel"
            description = "Panel with dashboard/panel links set"
          }
          links = [
            {
              name             = "Runbook"
              url              = "https://example.com/runbook"
              target_blank     = true
              render_variables = false
            }
          ]
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
                  ref = "#/spec/panels/33333333-3333-4333-8333-333333333333"
                }
              }
            ]
          }
        }
      }
    ]
  }
}
