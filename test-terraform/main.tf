terraform {
  required_providers {
    signoz = {
      source = "signoz.local/local/signoz"
      version = "0.0.1"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:8080"
  access_token = "VPMHMaDYFeYx5r4KXsr3AQ+7g7hazRjx7CtVt3X0jwQ="
}

# Example 1: High Memory Usage Alert
resource "signoz_alert" "high_memory" {
  alert            = "TF Test Alert - High Memory"
  alert_type       = "METRIC_BASED_ALERT"
  broadcast_to_all = false
  condition = jsonencode(
    {
      absentFor     = 10
      alertOnAbsent = true
      compositeQuery = {
        builderQueries = {
          A = {
            ShiftBy = 0
            aggregateAttribute = {
              dataType = "float64"
              isColumn = true
              isJSON   = false
              key      = "k8s_node_memory_rss"
              type     = "Gauge"
            }
            aggregateOperator = "avg"
            dataSource        = "metrics"
            disabled          = false
            expression        = "A"
            filters = {
              items = []
              op    = "AND"
            }
            groupBy = [
              {
                dataType = "string"
                isColumn = false
                isJSON   = false
                key      = "k8s_node_name"
                type     = "tag"
              },
            ]
            limit            = 0
            offset           = 0
            pageSize         = 0
            queryName        = "A"
            reduceTo         = "avg"
            spaceAggregation = "avg"
            stepInterval     = 60
            timeAggregation  = "avg"
          }
        }
        chQueries = {
          A = {
            disabled = false
            query    = ""
          }
        }
        panelType = "graph"
        promQueries = {
          A = {
            disabled = false
            query    = ""
          }
        }
        queryType = "builder"
        unit      = "bytes"
      }
      matchType         = "1"
      op                = "1"
      selectedQueryName = "A"
      target            = 10
      targetUnit        = "gbytes"
    }
  )
  description = "Alert is fired when the defined metric (current value: {{$value}}) crosses the threshold ({{$threshold}})"
  eval_window = "5m0s"
  frequency   = "1m0s"
  labels = {
    "observer" = "local-test"
  }
  preferred_channels = [
    "alert-test-terraform"
  ]
  rule_type = "threshold_rule"
  severity  = "info"
  version   = "v4"

  lifecycle {
    ignore_changes = [condition]
  }
}

# Example 2: Critical CPU Alert
resource "signoz_alert" "critical_cpu" {
  alert            = "TF Test Alert - Critical CPU"
  alert_type       = "METRIC_BASED_ALERT"
  condition = jsonencode({
    compositeQuery = {
      builderQueries = {
        A = {
          ShiftBy = 0
          aggregateAttribute = {
            dataType = "float64"
            isColumn = true
            isJSON   = false
            key      = "system_cpu_utilization"
            type     = "Gauge"
          }
          aggregateOperator = "avg"
          dataSource        = "metrics"
          disabled          = false
          expression        = "A"
          filters = {
            items = []
            op    = "AND"
          }
          groupBy          = []
          limit            = 0
          offset           = 0
          pageSize         = 0
          queryName        = "A"
          reduceTo         = "avg"
          spaceAggregation = "avg"
          stepInterval     = 60
          timeAggregation  = "avg"
        }
      }
      chQueries = {
        A = {
          disabled = false
          query    = ""
        }
      }
      panelType = "graph"
      promQueries = {
        A = {
          disabled = false
          query    = ""
        }
      }
      queryType = "builder"
    }
    op                = "1"
    target            = 80
    selectedQueryName = "A"
  })

  description = "CPU usage above 80%"
  summary     = "Critical CPU usage detected"
  eval_window = "13m0s"
  frequency   = "1m0s"
  disabled    = false
  rule_type   = "threshold_rule"
  severity    = "critical"
  version     = "v4"

  preferred_channels = [
    "alert-test-terraform"
  ]

  lifecycle {
    ignore_changes = [condition]
  }
}

# Example 3: Logs-based Alert with v2alpha1 Schema (Multi-threshold)
resource "signoz_alert" "logs_count_v2" {
  alert      = "Logs Count Alert (v2)"
  alert_type = "LOGS_BASED_ALERT"
  severity   = "critical"

  condition = jsonencode({
    compositeQuery = {
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            stepInterval = 0
            signal       = "logs"
            source       = ""
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
          }
        }
      ]
      panelType = "graph"
      queryType = "builder"
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name            = "critical"
          target          = 100
          targetUnit      = ""
          recoveryTarget  = null
          matchType       = "1"
          op              = "1"
          channels        = ["alert-test-terraform"]
        },
        {
          name            = "warning"
          target          = 50
          targetUnit      = ""
          recoveryTarget  = null
          matchType       = "1"
          op              = "1"
          channels        = ["alert-test-terraform"]
        }
      ]
    }
  })

  description      = "This alert is fired when log count crosses the threshold (current: {{$value}}, threshold: {{$threshold}})"
  summary          = "Log count alert triggered"
  eval_window      = "5m0s"
  frequency        = "1m0s"
  broadcast_to_all = false
  disabled         = false
  rule_type        = "threshold_rule"
  version          = "v5"

  schema_version = "v2alpha1"

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "15m0s"
      frequency  = "1m0s"
    }
  })

  notification_settings = {
    renotify = {
      interval     = "25m0s"
      alert_states = ["nodata","firing"]
      enabled      = true
    },
    group_by = ["asd"]
    use_policy = true
  }

  labels = {
    "team" = "platform"
  }

  lifecycle {
    ignore_changes = [condition]
  }
}

# Output the created alerts
output "high_memory_alert_id" {
  value = signoz_alert.high_memory.id
}

output "critical_cpu_alert_id" {
  value = signoz_alert.critical_cpu.id
}

output "logs_count_v2_alert_id" {
  value = signoz_alert.logs_count_v2.id
}

