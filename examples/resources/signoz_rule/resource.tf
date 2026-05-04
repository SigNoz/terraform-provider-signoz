terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint     = "http://localhost:8080"
  access_token = "<SIGNOZ-API-KEY>"
}

# Phase 2 minimum-viable rule:
#   alert_type = METRIC_BASED_ALERT
#   rule_type  = threshold_rule | promql_rule
#   query_type = promql, single query, no thresholds, no notification_settings
resource "signoz_rule" "high_error_rate" {
  alert       = "high-error-rate"
  alert_type  = "METRIC_BASED_ALERT"
  rule_type   = "promql_rule"
  version     = "v5"
  description = "Fires when 5xx rate over the last 5m exceeds 1 req/s."

  annotations = {
    summary     = "High error rate"
    description = "{{ $labels.service_name }} error rate is {{ $value }}"
  }

  labels = {
    severity = "warning"
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"
      queries = [{
        promql = {
          name  = "A"
          query = "rate(http_requests_total{status=~\"5..\"}[5m])"
          step  = "60s"
        }
      }]
    }
    target     = 1
    op         = "above"
    match_type = "at_least_once"
  }

  evaluation = {
    rolling = {
      eval_window = "5m"
      frequency   = "1m"
    }
  }

  preferred_channels = ["platform-alerts-slack"]
}

output "rule_id" {
  value = signoz_rule.high_error_rate.id
}
