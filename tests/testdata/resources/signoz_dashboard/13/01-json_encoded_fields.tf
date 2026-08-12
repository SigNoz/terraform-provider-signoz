# Every dashboard attribute the schema types as opaque JSON, set through
# jsonencode. There are 30 of them and only four distinct names --
# step_interval, functions[].args[].value, step and default_value -- repeated
# across mutually exclusive union arms, so the coverage comes from the spread of
# panels, queries and composite sub-queries rather than from the field count.
resource "signoz_dashboard" "json_encoded_fields" {
  schema_version = "v6"
  name           = "testdata-dashboard-json-encoded-fields-fyfgsb"
  tags = [
    {
      key   = "tag"
      value = "testdata"
    }
  ]

  spec = {
    display = {
      name        = "JSON-encoded fields"
      description = "Opaque-JSON attributes across every union arm that owns one: builder queries per signal, composite sub-queries, panel-level formula/promql/trace operator, and a list variable default."
    }
    links = []
    variables = [
      {
        list_variable = {
          kind = "ListVariable"
          spec = {
            name           = "service.name"
            allow_multiple = true
            default_value  = jsonencode(["frontend", "backend"])
            display = {
              name = "Service"
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
      "13000000-0000-4000-8000-000000000001" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Logs builder query"
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
                        signal        = "logs"
                        step_interval = jsonencode(60)
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
                        # args[].value is a number-or-string union; both arms
                        # survive a round-trip, unlike step_interval, which the
                        # server normalizes a duration string down to seconds.
                        functions = [
                          {
                            name = "cutoffmin"
                            args = [
                              {
                                name  = "threshold"
                                value = jsonencode(2)
                              }
                            ]
                          },
                          {
                            name = "timeshift"
                            args = [
                              {
                                name  = "shift"
                                value = jsonencode("60")
                              }
                            ]
                          }
                        ]
                        secondary_aggregations = [
                          {
                            expression    = "count()"
                            alias         = "logs_secondary"
                            step_interval = jsonencode(300)
                          }
                        ]
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
      "13000000-0000-4000-8000-000000000002" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Metrics builder query"
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
                        signal        = "metrics"
                        step_interval = jsonencode(60)
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
                        functions = [
                          {
                            name = "clampmax"
                            args = [
                              {
                                name  = "threshold"
                                value = jsonencode(3)
                              }
                            ]
                          }
                        ]
                        secondary_aggregations = [
                          {
                            expression    = "count()"
                            alias         = "metrics_secondary"
                            step_interval = jsonencode(300)
                          }
                        ]
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
      "13000000-0000-4000-8000-000000000003" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Traces builder query"
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
                        signal        = "traces"
                        step_interval = jsonencode(60)
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
                        functions = [
                          {
                            name = "clampmin"
                            args = [
                              {
                                name  = "threshold"
                                value = jsonencode(4)
                              }
                            ]
                          }
                        ]
                        secondary_aggregations = [
                          {
                            expression    = "count()"
                            alias         = "traces_secondary"
                            step_interval = jsonencode(300)
                          }
                        ]
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
      "13000000-0000-4000-8000-000000000004" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Composite query"
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
                  composite_query = {
                    kind = "signoz/CompositeQuery"
                    spec = {
                      queries = [
                        {
                          builder_query = {
                            type = "builder_query"
                            spec = {
                              logs = {
                                name          = "A"
                                signal        = "logs"
                                step_interval = jsonencode(60)
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
                                functions = [
                                  {
                                    name = "cutoffmin"
                                    args = [
                                      {
                                        name  = "threshold"
                                        value = jsonencode(6)
                                      }
                                    ]
                                  }
                                ]
                                secondary_aggregations = [
                                  {
                                    expression    = "count()"
                                    alias         = "composite_logs_secondary"
                                    step_interval = jsonencode(300)
                                  }
                                ]
                                legend = "composite logs"
                              }
                            }
                          }
                        },
                        {
                          builder_query = {
                            type = "builder_query"
                            spec = {
                              metrics = {
                                name          = "B"
                                signal        = "metrics"
                                step_interval = jsonencode(60)
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
                                functions = [
                                  {
                                    name = "clampmax"
                                    args = [
                                      {
                                        name  = "threshold"
                                        value = jsonencode(7)
                                      }
                                    ]
                                  }
                                ]
                                secondary_aggregations = [
                                  {
                                    expression    = "count()"
                                    alias         = "composite_metrics_secondary"
                                    step_interval = jsonencode(300)
                                  }
                                ]
                                legend = "composite metrics"
                              }
                            }
                          }
                        },
                        {
                          builder_query = {
                            type = "builder_query"
                            spec = {
                              traces = {
                                name          = "C"
                                signal        = "traces"
                                step_interval = jsonencode(60)
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
                                functions = [
                                  {
                                    name = "clampmin"
                                    args = [
                                      {
                                        name  = "threshold"
                                        value = jsonencode(8)
                                      }
                                    ]
                                  }
                                ]
                                secondary_aggregations = [
                                  {
                                    expression    = "count()"
                                    alias         = "composite_traces_secondary"
                                    step_interval = jsonencode(300)
                                  }
                                ]
                                legend = "composite traces"
                              }
                            }
                          }
                        },
                        {
                          builder_query = {
                            type = "builder_query"
                            spec = {
                              traces = {
                                name   = "D"
                                signal = "traces"
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
                                legend = "composite traces d"
                              }
                            }
                          }
                        },
                        {
                          builder_ai_query = {
                            type = "builder_ai_query"
                            spec = {
                              name          = "E"
                              signal        = "traces"
                              step_interval = jsonencode(60)
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
                              functions = [
                                {
                                  name = "absolute"
                                  args = [
                                    {
                                      name  = "threshold"
                                      value = jsonencode(9)
                                    }
                                  ]
                                }
                              ]
                              secondary_aggregations = [
                                {
                                  expression    = "count()"
                                  alias         = "composite_ai_secondary"
                                  step_interval = jsonencode(300)
                                }
                              ]
                              legend = "composite ai"
                            }
                          }
                        },
                        {
                          builder_formula = {
                            type = "builder_formula"
                            spec = {
                              name       = "F1"
                              expression = "A + B"
                              having = {
                                expression = ""
                              }
                              functions = [
                                {
                                  name = "clampmax"
                                  args = [
                                    {
                                      name  = "threshold"
                                      value = jsonencode(10)
                                    }
                                  ]
                                }
                              ]
                              legend = "composite formula"
                            }
                          }
                        },
                        {
                          builder_trace_operator = {
                            type = "builder_trace_operator"
                            spec = {
                              name          = "T1"
                              expression    = "C => D"
                              step_interval = jsonencode(60)
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
                              functions = [
                                {
                                  name = "cutoffmax"
                                  args = [
                                    {
                                      name  = "threshold"
                                      value = jsonencode(11)
                                    }
                                  ]
                                }
                              ]
                              legend = "composite trace operator"
                            }
                          }
                        },
                        {
                          promql = {
                            type = "promql"
                            spec = {
                              name   = "P1"
                              query  = "up"
                              step   = jsonencode(60)
                              stats  = false
                              legend = "composite promql"
                            }
                          }
                        }
                      ]
                    }
                  }
                }
              }
            }
          ]
        }
      }
      "13000000-0000-4000-8000-000000000005" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Panel-level formula"
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
                name = "F1"
                plugin = {
                  formula = {
                    kind = "signoz/Formula"
                    spec = {
                      name       = "F1"
                      expression = "A * 2"
                      having = {
                        expression = ""
                      }
                      functions = [
                        {
                          name = "clampmin"
                          args = [
                            {
                              name  = "threshold"
                              value = jsonencode(12)
                            }
                          ]
                        }
                      ]
                      legend = "formula"
                    }
                  }
                }
              }
            }
          ]
        }
      }
      "13000000-0000-4000-8000-000000000006" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Panel-level PromQL"
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
                  prom_qlquery = {
                    kind = "signoz/PromQLQuery"
                    spec = {
                      name   = "A"
                      query  = "up"
                      step   = jsonencode(60)
                      stats  = false
                      legend = "promql"
                    }
                  }
                }
              }
            }
          ]
        }
      }
      "13000000-0000-4000-8000-000000000007" = {
        kind = "Panel"
        spec = {
          display = {
            name = "Panel-level trace operator"
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
                name = "T1"
                plugin = {
                  trace_operator = {
                    kind = "signoz/TraceOperator"
                    spec = {
                      name          = "T1"
                      expression    = "A => B"
                      step_interval = jsonencode(60)
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
                      functions = [
                        {
                          name = "cutoffmax"
                          args = [
                            {
                              name  = "threshold"
                              value = jsonencode(5)
                            }
                          ]
                        }
                      ]
                      legend = "trace operator"
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
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000001"
                }
              },
              {
                x      = 0
                y      = 6
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000002"
                }
              },
              {
                x      = 0
                y      = 12
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000003"
                }
              },
              {
                x      = 0
                y      = 18
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000004"
                }
              },
              {
                x      = 0
                y      = 24
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000005"
                }
              },
              {
                x      = 0
                y      = 30
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000006"
                }
              },
              {
                x      = 0
                y      = 36
                width  = 12
                height = 6
                content = {
                  ref = "#/spec/panels/13000000-0000-4000-8000-000000000007"
                }
              }
            ]
          }
        }
      }
    ]
  }
}
