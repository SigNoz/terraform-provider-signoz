# Positive control for #154 ("dashboard create fails with invalid signal \"\"").
# One builder query per signal in one dashboard: each variant's own signal must
# reach the API. A metrics-only config cannot prove that — the generated union
# builder pins signal to the variant's discriminator, so "metrics" would survive
# even if the configured value were ignored. logs and traces are what discriminate.
resource "signoz_dashboard" "builder_query_signals" {
  schema_version = "v6"
  name           = "testdata-dashboard-builder-query-signals-tk52wr"
  tags = [
    {
      key   = "tag"
      value = "testdata"
    }
  ]

  spec = {
    display = {
      name        = "Builder query signals"
      description = "A logs, a metrics and a traces builder query side by side; signal must round-trip per variant."
    }
    links     = []
    variables = []
    panels = {
      "00000000-0000-4000-8000-000000000111" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Log volume"
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
                    spec = {
                      logs = {
                        name          = "A"
                        step_interval = "60"
                        signal        = "logs"
                        aggregations = [
                          {
                            expression = "count()"
                          }
                        ]
                        filter = {
                          expression = ""
                        }
                        having = {
                          expression = ""
                        }
                        legend = "logs"
                      }
                    }
                  }
                }
              }
            }
          ]
        }
      }
      "00000000-0000-4000-8000-000000000112" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Memory usage"
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
                  unit              = "bytes"
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
                        legend = "metrics"
                      }
                    }
                  }
                }
              }
            }
          ]
        }
      }
      "00000000-0000-4000-8000-000000000113" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Span count"
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
                    spec = {
                      traces = {
                        name          = "A"
                        step_interval = "60"
                        signal        = "traces"
                        aggregations = [
                          {
                            expression = "count()"
                          }
                        ]
                        filter = {
                          expression = ""
                        }
                        having = {
                          expression = ""
                        }
                        legend = "traces"
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
                  ref = "#/spec/panels/00000000-0000-4000-8000-000000000111"
                }
              },
              {
                x      = 0
                y      = 6
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/00000000-0000-4000-8000-000000000112"
                }
              },
              {
                x      = 0
                y      = 12
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/00000000-0000-4000-8000-000000000113"
                }
              }
            ]
          }
        }
      }
    ]
  }
}
