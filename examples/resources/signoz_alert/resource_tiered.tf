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
