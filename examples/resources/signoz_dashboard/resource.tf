# A SigNoz dashboard (v2, Perses-based): a Redis overview with a host filter
# variable and three panels — an ungrouped time series, a time series grouped by
# host, and a single-stat number panel. See the schema below for all attributes.

resource "signoz_dashboard" "redis_overview" {
  schema_version = "v6"
  name           = "redis-overview-aqdiza77"
  tags = [
    {
      key   = "tag"
      value = "redis"
    },
    {
      key   = "tag"
      value = "database"
    },
  ]
  spec = {
    display = {
      name        = "Redis overview"
      description = "This dashboard shows the Redis instance overview. It includes latency, hit/miss rate, connections, and memory information."
    }
    links = []
    variables = [
      {
        list_variable = {
          kind = "ListVariable"
          spec = {
            display = {
              name        = "host_name"
              description = "List of hosts sending Redis metrics"
            }
            allow_all_value = true
            allow_multiple  = true
            sort            = "alphabetical-asc"
            name            = "host_name"
            plugin = {
              dynamic_variable = {
                kind = "signoz/DynamicVariable"
                spec = {
                  name   = "host.name"
                  signal = "metrics"
                }
              }
            }
          }
        }
      },
    ]
    panels = {
      "2fbaef0d-3cdb-4ce3-aa3c-9bbbb41786d9" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Command/s"
          }
          links = []
          plugin = {
            time_series_panel = {
              kind = "signoz/TimeSeriesPanel"
              spec = {
                visualization = {
                  time_preference = "global_time"
                  fill_spans      = true
                }
                formatting = {
                  unit              = "ops"
                  decimal_precision = "2"
                }
                chart_appearance = {
                  line_interpolation = "spline"
                  show_points        = false
                  line_style         = "solid"
                  fill_mode          = "solid"
                  span_gaps = {
                    fill_only_below = false
                    fill_less_than  = "0s"
                  }
                }
                axes = {
                  soft_min     = 0
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
                            metric_name       = "redis.commands"
                            time_aggregation  = "avg"
                            space_aggregation = "sum"
                            reduce_to         = "sum"
                          },
                        ]
                        filter = {
                          expression = "host.name IN $host_name"
                        }
                        having = {
                          expression = ""
                        }
                        legend = "ops/s"
                      }
                    }
                  }
                }
              }
            },
          ]
        }
      }
      "9698cee2-b1f3-4c0b-8c9f-3da4f0e05f17" = {
        kind = "Panel"
        spec = {
          display = {
            name = "RSS Memory"
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
                chart_appearance = {
                  line_interpolation = "spline"
                  show_points        = false
                  line_style         = "solid"
                  fill_mode          = "solid"
                  span_gaps = {
                    fill_only_below = false
                    fill_less_than  = "0s"
                  }
                }
                axes = {
                  soft_min     = 0
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
                            metric_name       = "redis.memory.rss"
                            time_aggregation  = "avg"
                            space_aggregation = "sum"
                            reduce_to         = "sum"
                          },
                        ]
                        filter = {
                          expression = "host.name IN $host_name"
                        }
                        group_by = [
                          {
                            name            = "host.name"
                            field_context   = "attribute"
                            field_data_type = "string"
                          },
                        ]
                        having = {
                          expression = ""
                        }
                        legend = "Rss::{{host.name}}"
                      }
                    }
                  }
                }
              }
            },
          ]
        }
      }
      "f6a7b8c9-d0e1-4f2a-c13d-4e5f6a7b8c9d" = {
        kind = "Panel"
        spec = {
          display = {
            name        = "Uptime"
            description = "Number of seconds since Redis server start"
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
                            metric_name       = "redis.uptime"
                            time_aggregation  = "max"
                            space_aggregation = "max"
                            reduce_to         = "last"
                          },
                        ]
                        filter = {
                          expression = "host.name IN $host_name"
                        }
                        having = {
                          expression = ""
                        }
                        legend = "Uptime"
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
            display = {
              title = "Overview"
              collapse = {
                open = true
              }
            }
            items = [
              {
                x      = 0
                y      = 0
                width  = 8
                height = 6
                content = {
                  ref = "#/spec/panels/2fbaef0d-3cdb-4ce3-aa3c-9bbbb41786d9"
                }
              },
              {
                x      = 8
                y      = 0
                width  = 4
                height = 6
                content = {
                  ref = "#/spec/panels/f6a7b8c9-d0e1-4f2a-c13d-4e5f6a7b8c9d"
                }
              },
              {
                x      = 0
                y      = 6
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/9698cee2-b1f3-4c0b-8c9f-3da4f0e05f17"
                }
              },
            ]
          }
        }
      },
    ]
  }
}
