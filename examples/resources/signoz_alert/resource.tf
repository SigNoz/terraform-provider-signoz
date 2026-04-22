terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:3301"
  access_token = "<SIGNOZ-API-KEY>"
}

# Metric threshold (single builder query).
# Fires when a pod consumes more than 80% of its requested CPU for the whole
# evaluation window. Uses the k8s.pod.cpu_request_utilization metric.
resource "signoz_alert" "metric_threshold_single" {
  alert       = "Pod CPU above 80% of request"
  alert_type  = "METRIC_BASED_ALERT"
  description = "CPU usage for api-service pods exceeds 80% of the requested CPU"
  summary     = "Pod CPU above {{$threshold}} of request"
  severity    = "critical"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "percentunit"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "metrics"
            stepInterval = 60
            aggregations = [
              {
                metricName       = "k8s.pod.cpu_request_utilization"
                timeAggregation  = "avg"
                spaceAggregation = "max"
              }
            ]
            filter = {
              expression = "k8s.deployment.name = 'api-service'"
            }
            groupBy = [
              { name = "k8s.pod.name", fieldContext = "resource", fieldDataType = "string" },
              { name = "deployment.environment", fieldContext = "resource", fieldDataType = "string" },
            ]
            legend = "{{k8s.pod.name}} ({{deployment.environment}})"
          }
        }
      ]
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name      = "critical"
          op        = "above"
          matchType = "all_the_times"
          target    = 0.8
          channels  = ["slack-platform", "pagerduty-oncall"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "15m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["k8s.pod.name", "deployment.environment"]
    renotify = {
      enabled      = true
      interval     = "4h"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "platform"
  }
}

# Logs threshold with builder formula (error rate percentage).
# Two disabled log-count queries (A = errors, B = total) combined via a
# builder_formula into a percentage — the classic service-level error-rate shape.
resource "signoz_alert" "logs_error_rate_formula" {
  alert       = "Payments-api error log rate above 1%"
  alert_type  = "LOGS_BASED_ALERT"
  description = "Error log ratio as a percentage of total logs for payments-api"
  summary     = "Payments-api error rate above {{$threshold}}%"
  severity    = "critical"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "percent"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "logs"
            stepInterval = 60
            disabled     = true
            aggregations = [{ expression = "count()" }]
            filter = {
              expression = "service.name = 'payments-api' AND severity_text IN ['ERROR', 'FATAL']"
            }
            groupBy = [
              { name = "deployment.environment", fieldContext = "resource", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_query"
          spec = {
            name         = "B"
            signal       = "logs"
            stepInterval = 60
            disabled     = true
            aggregations = [{ expression = "count()" }]
            filter       = { expression = "service.name = 'payments-api'" }
            groupBy = [
              { name = "deployment.environment", fieldContext = "resource", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_formula"
          spec = {
            name       = "F1"
            expression = "(A / B) * 100"
            legend     = "{{deployment.environment}}"
          }
        }
      ]
    }
    selectedQueryName = "F1"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name      = "critical"
          op        = "above"
          matchType = "at_least_once"
          target    = 1
          channels  = ["slack-payments"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "5m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["deployment.environment"]
    renotify = {
      enabled      = true
      interval     = "30m"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "payments"
  }
}

# Traces threshold — p99 latency with ns → s unit conversion.
# The series unit is ns (compositeQuery.unit) but the target is in seconds
# (threshold.targetUnit). SigNoz converts before comparing.
resource "signoz_alert" "traces_threshold_latency" {
  alert       = "Search API p99 latency above 5s"
  alert_type  = "TRACES_BASED_ALERT"
  description = "p99 duration of the search endpoint exceeds 5s"
  summary     = "Search-api latency degraded"
  severity    = "warning"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "ns"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "traces"
            stepInterval = 60
            aggregations = [{ expression = "p99(duration_nano)" }]
            filter = {
              expression = "service.name = 'search-api' AND name = 'GET /api/v1/search'"
            }
            groupBy = [
              { name = "service.name", fieldContext = "resource", fieldDataType = "string" },
              { name = "http.route", fieldContext = "attribute", fieldDataType = "string" },
            ]
            legend = "{{service.name}} {{http.route}}"
          }
        }
      ]
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name       = "warning"
          op         = "above"
          matchType  = "at_least_once"
          target     = 5
          targetUnit = "s"
          channels   = ["slack-search"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "5m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["service.name", "http.route"]
    renotify = {
      enabled      = true
      interval     = "30m"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "search"
  }
}

# Tiered thresholds — warning and critical in a single rule with different
# targets and channels, plus absent-data handling.
resource "signoz_alert" "tiered_thresholds" {
  alert       = "Kafka consumer lag warn / critical"
  alert_type  = "METRIC_BASED_ALERT"
  description = "Warn at lag >= 50 and page at >= 200, tiered via thresholds.spec."
  summary     = "Kafka consumer lag"
  severity    = "critical"
  rule_type   = "threshold_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      queries = [
        {
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "metrics"
            stepInterval = 60
            disabled     = true
            aggregations = [
              { metricName = "kafka_log_end_offset", timeAggregation = "max", spaceAggregation = "max" },
            ]
            filter = { expression = "topic != '__consumer_offsets'" }
            groupBy = [
              { name = "topic", fieldContext = "attribute", fieldDataType = "string" },
              { name = "partition", fieldContext = "attribute", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_query"
          spec = {
            name         = "B"
            signal       = "metrics"
            stepInterval = 60
            disabled     = true
            aggregations = [
              { metricName = "kafka_consumer_committed_offset", timeAggregation = "max", spaceAggregation = "max" },
            ]
            filter = { expression = "topic != '__consumer_offsets'" }
            groupBy = [
              { name = "topic", fieldContext = "attribute", fieldDataType = "string" },
              { name = "partition", fieldContext = "attribute", fieldDataType = "string" },
            ]
          }
        },
        {
          type = "builder_formula"
          spec = {
            name       = "F1"
            expression = "A - B"
            legend     = "{{topic}}/{{partition}}"
          }
        }
      ]
    }
    alertOnAbsent     = true
    absentFor         = 15
    selectedQueryName = "F1"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name      = "warning"
          op        = "above"
          matchType = "all_the_times"
          target    = 50
          channels  = ["slack-kafka-info"]
        },
        {
          name      = "critical"
          op        = "above"
          matchType = "all_the_times"
          target    = 200
          channels  = ["slack-kafka-alerts", "pagerduty-kafka"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "5m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["topic"]
    renotify = {
      enabled      = true
      interval     = "15m"
      alert_states = ["firing", "nodata"]
    }
    use_policy = false
  }

  labels = {
    team = "data-platform"
  }
}

# PromQL rule — use when a query is easier to express in PromQL than in the
# builder. Dotted OTEL resource attributes are quoted in the expression.
resource "signoz_alert" "metric_promql" {
  alert       = "Kafka consumer group lag above 1000"
  alert_type  = "METRIC_BASED_ALERT"
  description = "Consumer group lag computed via PromQL"
  summary     = "Kafka consumer lag high"
  severity    = "critical"
  rule_type   = "promql_rule"
  version     = "v5"

  schema_version = "v2alpha1"

  condition = jsonencode({
    compositeQuery = {
      queryType = "promql"
      panelType = "graph"
      queries = [
        {
          type = "promql"
          spec = {
            name   = "A"
            query  = "(max by(topic, partition, \"deployment.environment\")(kafka_log_end_offset) - on(topic, partition, \"deployment.environment\") group_right max by(group, topic, partition, \"deployment.environment\")(kafka_consumer_committed_offset)) > 0"
            legend = "{{topic}}/{{partition}} ({{group}})"
          }
        }
      ]
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [
        {
          name      = "critical"
          op        = "above"
          matchType = "all_the_times"
          target    = 1000
          channels  = ["slack-data-platform", "pagerduty-data"]
        }
      ]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = {
      evalWindow = "10m"
      frequency  = "1m"
    }
  })

  notification_settings = {
    group_by = ["group", "topic"]
    renotify = {
      enabled      = true
      interval     = "1h"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "data-platform"
  }
}

output "alert_new" {
  value = signoz_alert.metric_threshold_single
}
