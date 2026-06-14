# Metric anomaly rule.
# Wraps a builder query in the `anomaly` function with daily seasonality;
# SigNoz compares each point against the forecast for that time of day and
# fires when the z-score stays below the threshold for the entire window.
# `requireMinPoints` guards against noisy intervals.
resource "signoz_alert" "metric_anomaly" {
  alert       = "Anomalous drop in ingested spans"
  alert_type  = "METRIC_BASED_ALERT"
  description = "Detect an abrupt drop in span ingestion using a z-score anomaly function"
  summary     = "Span ingestion anomaly"
  severity    = "warning"
  rule_type   = "anomaly_rule"
  version     = "v5"

  eval_window = "24h"
  frequency   = "3h"

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
            stepInterval = 21600
            aggregations = [
              { metricName = "otelcol_receiver_accepted_spans", timeAggregation = "rate", spaceAggregation = "sum" },
            ]
            filter = { expression = "tenant_tier = 'premium'" }
            groupBy = [
              { name = "tenant_id", fieldContext = "attribute", fieldDataType = "string" },
            ]
            functions = [
              {
                name = "anomaly"
                args = [{ name = "z_score_threshold", value = 2 }]
              }
            ]
            legend = "{{tenant_id}}"
          }
        }
      ]
    }
    op                = "below"
    matchType         = "all_the_times"
    target            = 2
    algorithm         = "standard"
    seasonality       = "daily"
    selectedQueryName = "A"
    requireMinPoints  = true
    requiredNumPoints = 3
  })

  preferred_channels = ["slack-ingestion"]

  labels = {
    severity = "warning"
  }
}
