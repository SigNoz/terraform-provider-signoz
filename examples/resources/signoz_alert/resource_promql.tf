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
          channels  = ["slack", "pagerduty"]
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
