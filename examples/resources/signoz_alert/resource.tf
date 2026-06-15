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
          channels  = ["slack", "pagerduty"]
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
