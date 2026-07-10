# Minimal rule: only the attributes the schema marks required, plus the query
# body the server needs to parse the payload. Every optional field (evaluation,
# thresholds, notification_settings, labels, annotations, description, ...) is
# omitted. Uses a PromQL query, the simplest query kind (no builder signal).
resource "signoz_rule" "minimal" {
  alert      = "minimal-required-only"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "promql_rule"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"

      queries = [
        {
          promql = {
            type = "promql"
            spec = {
              name  = "A"
              query = "up"
            }
          }
        }
      ]
    }
  }
}
