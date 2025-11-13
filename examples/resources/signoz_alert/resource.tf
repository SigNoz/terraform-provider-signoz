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
            IsAnomaly= false
            QueriesUsedInFormula = null
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

resource "signoz_alert" "new_alert_v2" {
  alert      = "new alert with v2 schema"
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
          name           = "critical"
          target         = 100
          targetUnit     = ""
          recoveryTarget = null
          matchType      = "1"
          op             = "1"
          channels       = ["alert-test-terraform"]
        },
        {
          name           = "warning"
          target         = 50
          targetUnit     = ""
          recoveryTarget = null
          matchType      = "1"
          op             = "1"
          channels       = ["alert-test-terraform"]
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
      evalWindow = "35m0s"
      frequency  = "1m0s"
    }
  })

  notification_settings = {
    renotify = {
      interval     = "25m0s"
      alert_states = ["nodata", "firing"]
      enabled      = true
    },
    group_by   = ["container.id"]
    use_policy = true
  }

  preferred_channels = ["alert-test-terraform"]

  labels = {
    "team" = "platform"
  }
}

output "alert_new" {
  value = signoz_alert.new_alert
}
