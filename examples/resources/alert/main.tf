terraform {
  required_providers {
    signoz = {
      source = "registry.terraform.io/signoz/signoz"
    }
  }
}

provider "signoz" {
  endpoint = "http://localhost:3301"
  # access_token = "ACCESS_TOKEN"
}

resource "signoz_alert" "new_alert" {
  alert            = "TF Test Alert"
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
}

output "alert_new" {
  value = signoz_alert.new_alert
}
