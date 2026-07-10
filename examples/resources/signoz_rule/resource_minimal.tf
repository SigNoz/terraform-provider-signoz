# Minimal rule: only the attributes the schema marks required. Every optional
# field (evaluation, thresholds, notification_settings, labels, annotations,
# description, ...) is omitted.
resource "signoz_rule" "minimal" {
  alert      = "minimal-required-only"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"

      queries = [
        {
          builder_query = {
            type = "builder_query"
          }
        }
      ]
    }
  }
}
