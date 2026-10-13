# Using the SigNoz Terraform Provider

This guide covers provider setup, authentication, and usage of all managed resources and data sources.

## Provider Configuration

```hcl
terraform {
  required_providers {
    signoz = {
      source  = "registry.terraform.io/signoz/signoz"
      version = ">= 0.1.0"
    }
  }
}

provider "signoz" {
  endpoint     = "https://signoz.example.com"
  access_token = var.signoz_access_token
}
```

### Configuration Reference

| Attribute | Env Variable | Default | Description |
|-----------|-------------|---------|-------------|
| `access_token` | `SIGNOZ_ACCESS_TOKEN` | *(required)* | API access token (Admin role). Retrieve from SigNoz UI or create a Service Account. |
| `endpoint` | `SIGNOZ_ENDPOINT` | `http://localhost:3301` | Root URL of SigNoz UI |
| `http_timeout` | `SIGNOZ_HTTP_TIMEOUT` | `35` | Per-request timeout in seconds |
| `http_max_retry` | `SIGNOZ_HTTP_MAX_RETRY` | `10` | Max retry count with constant backoff |

HCL values override env vars, which override defaults.

### Authentication

Generate an access token in SigNoz UI under **Settings > Access Tokens** (requires Admin role), or create a Service Account with API keys. Pass via environment variable to avoid committing secrets:

```bash
export SIGNOZ_ACCESS_TOKEN="your-token-here"
export SIGNOZ_ENDPOINT="https://signoz.example.com"
terraform plan
```

---

## Resources

### signoz_alert

Manages alert rules. Supports metric, log, trace, and exception-based alerts with threshold or PromQL rules.

```hcl
resource "signoz_alert" "high_latency" {
  alert      = "High API Latency"
  alert_type = "METRIC_BASED_ALERT"
  severity   = "warning"
  rule_type  = "threshold_rule"
  version    = "v4"

  condition = jsonencode({
    compositeQuery = {
      builderQueries = {
        A = {
          dataSource  = "metrics"
          queryName   = "A"
          aggregateAttribute = { key = "signoz_latency_bucket" }
          aggregateOperator  = "p99"
        }
      }
      queryType = "builder"
    }
    op        = "1"
    target    = 500
    matchType = "4"
  })

  eval_window        = "5m0s"
  frequency          = "1m0s"
  labels             = { team = "platform" }
  preferred_channels = ["slack-alerts"]
}
```

**Schema version v2** enables `notification_settings` and `evaluation` blocks:

```hcl
resource "signoz_alert" "critical_errors" {
  alert          = "Critical Error Rate"
  alert_type     = "LOGS_BASED_ALERT"
  severity       = "critical"
  schema_version = "v2alpha1"
  version        = "v5"
  rule_type      = "threshold_rule"
  condition      = jsonencode({ /* ... */ })

  evaluation = jsonencode({
    kind = "rolling"
    spec = { evalWindow = "15m0s", frequency = "1m0s" }
  })

  notification_settings = {
    renotify = {
      interval     = "30m0s"
      alert_states = ["firing", "nodata"]
      enabled      = true
    }
    group_by   = ["service.name"]
    use_policy = true
  }
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `alert` | String | yes | | Alert name |
| `alert_type` | String | yes | | `METRIC_BASED_ALERT`, `LOGS_BASED_ALERT`, `TRACES_BASED_ALERT`, `EXCEPTIONS_BASED_ALERT` |
| `severity` | String | yes | | `info`, `warning`, `error`, `critical` |
| `condition` | String | yes | | Alert condition (JSON) |
| `rule_type` | String | | | `threshold_rule` or `promql_rule` |
| `eval_window` | String | | `5m0s` | Evaluation window |
| `frequency` | String | | `1m0s` | Evaluation frequency |
| `version` | String | | `v4` | Alert version |
| `schema_version` | String | | | `v1` or `v2alpha1` (v2 enables notification_settings/evaluation) |
| `disabled` | Bool | | `false` | Disable the alert |
| `labels` | Map(String) | | | Custom labels |
| `preferred_channels` | List(String) | | | Channel names to send alerts to |
| `notification_settings` | Object | | | v2 only: renotify, group_by, use_policy |
| `evaluation` | String | | | v2 only: evaluation config (JSON) |
| `id` | String | computed | | Alert ID |
| `state` | String | computed | | `inactive`, `pending`, `firing`, `disabled`, `nodata` |

Import: `terraform import signoz_alert.example <alert-id>`

---

### signoz_dashboard

Manages dashboards. Layout, widgets, variables, and panel_map are stored as JSON strings.

```hcl
resource "signoz_dashboard" "api_overview" {
  title                     = "API Overview"
  name                      = "api-overview"
  description               = "API latency and error rate"
  version                   = "v4"
  collapsable_rows_migrated = true
  uploaded_grafana          = false
  tags                      = ["api", "production"]

  layout    = jsonencode([{ h = 8, i = "widget-1", w = 6, x = 0, y = 0 }])
  widgets   = jsonencode([{ /* widget config */ }])
  variables = jsonencode({})
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | String | yes | Dashboard title |
| `name` | String | yes | Dashboard name |
| `description` | String | yes | Description |
| `version` | String | yes | Version (e.g. `v4`) |
| `collapsable_rows_migrated` | Bool | yes | Collapsable rows migration flag |
| `uploaded_grafana` | Bool | yes | Grafana import flag |
| `layout` | String | yes | Layout config (JSON) |
| `widgets` | String | yes | Widgets config (JSON) |
| `variables` | String | yes | Variables config (JSON) |
| `panel_map` | String | | Panel map (JSON) |
| `tags` | List(String) | | Tags |
| `id` | String | computed | Dashboard UUID |

Import: `terraform import signoz_dashboard.example <dashboard-uuid>`

---

### signoz_notification_channel

Manages notification channels. Alerts reference channels by ID. Each channel has a `type` that determines which config block to use.

```hcl
# Slack channel
resource "signoz_notification_channel" "slack" {
  name = "slack-alerts"
  type = "slack"

  slack_configs = {
    api_url = "https://hooks.example.com/services/TXXXXXXXXX/BXXXXXXXXX/xxxxxxxx"
    channel = "#alerts"
    title   = "SigNoz Alert"
    text    = "{{ .CommonLabels.alertname }}: {{ .CommonAnnotations.description }}"
  }
}

# Webhook channel
resource "signoz_notification_channel" "webhook" {
  name = "incident-webhook"
  type = "webhook"

  webhook_configs = {
    api_url  = "https://incident.example.com/api/webhook"
    username = "signoz"
    password = var.webhook_password
  }
}

# PagerDuty channel
resource "signoz_notification_channel" "pagerduty" {
  name = "pagerduty-critical"
  type = "pagerduty"

  pagerduty_configs = {
    routing_key = var.pagerduty_routing_key
  }
}

# Email channel
resource "signoz_notification_channel" "email" {
  name = "ops-email"
  type = "email"

  email_configs = {
    to            = "ops@example.com"
    send_resolved = true
  }
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | yes | Channel name |
| `type` | String | yes | `slack`, `webhook`, `pagerduty`, `opsgenie`, `msteams`, `email` |
| `slack_configs` | Object | | Slack: `api_url` (required), `channel`, `title`, `text` |
| `webhook_configs` | Object | | Webhook: `api_url` (required), `username`, `password` (sensitive) |
| `pagerduty_configs` | Object | | PagerDuty: `routing_key` (sensitive), `service_key` (sensitive), `details` |
| `opsgenie_configs` | Object | | OpsGenie: `api_key` (required, sensitive), `api_url` |
| `msteams_configs` | Object | | MS Teams: `webhook_url` (required) |
| `email_configs` | Object | | Email: `to` (required), `send_resolved`, `html`, `headers` |
| `id` | String | computed | Channel ID |

Provide exactly one config block matching the `type`.

Import: `terraform import signoz_notification_channel.example <channel-id>`

---

### signoz_route_policy

Manages alert routing policies. Routes matching alerts to specific notification channels.

```hcl
resource "signoz_route_policy" "critical_to_pagerduty" {
  route_name  = "critical-alerts"
  description = "Route critical alerts to PagerDuty"
  enabled     = true

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }

  channel_ids = [signoz_notification_channel.pagerduty.id]

  group_wait      = "30s"
  group_interval  = "5m"
  repeat_interval = "4h"
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `route_name` | String | yes | | Policy name |
| `matchers` | List(Object) | yes | | Match rules: `match_name`, `match_value`, `match_type` |
| `channel_ids` | List(String) | yes | | Target notification channel IDs |
| `enabled` | Bool | | `true` | Enable the policy |
| `continue_matching` | Bool | | `false` | Continue to next matching policy |
| `group_wait` | String | | `30s` | Delay before first notification |
| `group_interval` | String | | `5m` | Interval between group notifications |
| `repeat_interval` | String | | `4h` | Re-notification interval |
| `description` | String | | | Description |
| `id` | String | computed | | Policy ID |

Import: `terraform import signoz_route_policy.example <policy-id>`

---

### signoz_downtime_schedule

Manages planned maintenance windows that suppress alert notifications.

```hcl
# Weekly maintenance window
resource "signoz_downtime_schedule" "weekly_maintenance" {
  name        = "Weekly Maintenance"
  description = "Suppress alerts during weekly maintenance"
  all_alerts  = true

  schedule {
    start_time = "2024-01-15T02:00:00Z"
    end_time   = "2024-01-15T04:00:00Z"
    timezone   = "America/New_York"

    recurrence {
      type         = "weekly"
      days_of_week = [0, 6]
      duration     = "2h"
    }
  }
}

# One-time downtime for specific alerts
resource "signoz_downtime_schedule" "deploy_window" {
  name      = "Production Deploy"
  alert_ids = [signoz_alert.high_latency.id]

  schedule {
    start_time = "2024-03-15T14:00:00Z"
    end_time   = "2024-03-15T16:00:00Z"
    timezone   = "UTC"

    recurrence {
      type = "once"
    }
  }
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | String | yes | | Schedule name |
| `schedule` | Object | yes | | Schedule config: `start_time`, `end_time`, `timezone`, optional `recurrence` |
| `schedule.recurrence` | Object | | | `type` (`once`/`daily`/`weekly`), `days_of_week` (0-6), `duration` |
| `alert_ids` | List(String) | | | Specific alert IDs to suppress |
| `all_alerts` | Bool | | `false` | Suppress all alerts |
| `description` | String | | | Description |
| `id` | String | computed | | Schedule ID |

Import: `terraform import signoz_downtime_schedule.example <schedule-id>`

---

### signoz_data_retention

Manages global data retention settings. This is a **singleton resource** — only one instance should exist per SigNoz deployment. Destroying this resource removes it from Terraform state only; retention settings persist in SigNoz.

```hcl
resource "signoz_data_retention" "this" {
  metrics_ttl_duration_hrs      = 720   # 30 days
  metrics_move_ttl_duration_hrs = 0     # cold storage disabled
  logs_ttl_duration_hrs         = 360   # 15 days
  logs_move_ttl_duration_hrs    = 0
  traces_ttl_duration_hrs       = 168   # 7 days
  traces_move_ttl_duration_hrs  = 0
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `metrics_ttl_duration_hrs` | Int64 | | `720` | Metrics retention (hours) |
| `metrics_move_ttl_duration_hrs` | Int64 | | `0` | Metrics cold storage threshold (0 = disabled) |
| `logs_ttl_duration_hrs` | Int64 | | `720` | Logs retention (hours) |
| `logs_move_ttl_duration_hrs` | Int64 | | `0` | Logs cold storage threshold |
| `traces_ttl_duration_hrs` | Int64 | | `720` | Traces retention (hours) |
| `traces_move_ttl_duration_hrs` | Int64 | | `0` | Traces cold storage threshold |
| `id` | String | computed | `settings` | Always `"settings"` |

Import: `terraform import signoz_data_retention.this settings`

---

### signoz_saved_view

Manages saved explorer views for logs, traces, and metrics.

```hcl
resource "signoz_saved_view" "error_logs" {
  name        = "Error Logs - Production"
  source_page = "logs"
  tags        = ["production", "errors"]

  explorer_data = jsonencode({
    query = {
      queryType = "builder"
      builder = {
        queryData = [{
          dataSource = "logs"
          queryName  = "A"
          filters = {
            items = [{
              key   = { key = "severity_text", dataType = "string", type = "tag" }
              op    = "="
              value = "ERROR"
            }]
            op = "AND"
          }
        }]
      }
    }
  })
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | yes | View name |
| `source_page` | String | yes | `logs`, `traces`, or `metrics` |
| `explorer_data` | String | yes | Explorer query config (JSON) |
| `tags` | List(String) | | Tags |
| `category` | String | | Category |
| `extra_data` | String | | Additional data |
| `uuid` | String | computed | View UUID (used as import ID) |

Import: `terraform import signoz_saved_view.example <view-uuid>`

---

### signoz_ingestion_key

Manages ingestion API keys with optional rate limits. The `value` field (the actual API key) is sensitive and only returned on creation — it is preserved in Terraform state across subsequent reads.

```hcl
resource "signoz_ingestion_key" "production" {
  name = "Production Ingest Key"
  tags = ["production", "team-platform"]

  rate_limits = {
    day_limit    = 1000000
    second_limit = 1000
  }
}

output "ingestion_url" {
  value = signoz_ingestion_key.production.ingestion_url
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | yes | Key name |
| `tags` | List(String) | | Tags |
| `expires_at` | String | | Expiration time |
| `rate_limits` | Object | | `day_limit` (Int64), `second_limit` (Int64) |
| `id` | String | computed | Key ID |
| `value` | String | computed, sensitive | The API key (only returned on creation) |
| `ingestion_url` | String | computed | Ingestion endpoint URL |

Import: `terraform import signoz_ingestion_key.example <key-id>`

**Note:** Importing an existing key will populate all fields except `value`, which will be empty since the API does not return it on GET requests.

---

### signoz_log_pipeline

Manages log processing pipelines. Pipelines transform log data using configurable filters and processors.

**Important:** The SigNoz log pipeline API uses a batch-replace pattern. The provider handles this transparently — create, update, and delete operations perform a read-modify-write cycle internally.

```hcl
resource "signoz_log_pipeline" "extract_severity" {
  name        = "Extract Severity"
  alias       = "extract-severity"
  description = "Extract severity level from log body"
  enabled     = true
  order_id    = 1

  filter = jsonencode({
    op    = "AND"
    items = [{
      key = { key = "service_name", type = "tag" }
      op  = "="
      value = "api-server"
    }]
  })

  config = jsonencode([{
    id         = "proc-1"
    orderId    = 1
    enabled    = true
    name       = "Extract severity"
    type       = "grok_parser"
    output     = "severity"
    parse_from = "body"
    pattern    = "%{WORD:severity}"
  }])
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | String | yes | | Pipeline name |
| `order_id` | Int64 | yes | | Processing order (lower = earlier) |
| `filter` | String | yes | | Filter config (JSON) |
| `config` | String | yes | | Processor config (JSON array) |
| `alias` | String | | | Alias identifier |
| `description` | String | | | Description |
| `enabled` | Bool | | `true` | Enable the pipeline |
| `id` | String | computed | | Pipeline ID |

Import: `terraform import signoz_log_pipeline.example <pipeline-id>`

---

## Data Sources

Every resource has a matching data source for read-only lookups by ID:

```hcl
data "signoz_alert" "existing"                { id = "5" }
data "signoz_dashboard" "existing"            { id = "abc-123" }
data "signoz_notification_channel" "existing"  { id = "1" }
data "signoz_route_policy" "existing"          { id = "1" }
data "signoz_downtime_schedule" "existing"     { id = "1" }
data "signoz_data_retention" "current"         {}
data "signoz_saved_view" "existing"            { uuid = "view-uuid-here" }
data "signoz_ingestion_key" "existing"         { id = "key-id-here" }
data "signoz_log_pipeline" "existing"          { id = "pipeline-id-here" }
```

All attributes are computed in data sources. The `signoz_data_retention` data source requires no input (singleton).

---

## Common Patterns

### Full alerting pipeline

Wire channels, routing, and alerts together:

```hcl
resource "signoz_notification_channel" "slack" {
  name = "slack-ops"
  type = "slack"
  slack_configs = {
    api_url = var.slack_webhook_url
    channel = "#ops-alerts"
  }
}

resource "signoz_notification_channel" "pagerduty" {
  name = "pagerduty-critical"
  type = "pagerduty"
  pagerduty_configs = {
    routing_key = var.pagerduty_key
  }
}

resource "signoz_route_policy" "critical" {
  route_name = "critical-to-pagerduty"
  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }
  channel_ids = [signoz_notification_channel.pagerduty.id]
}

resource "signoz_route_policy" "default" {
  route_name = "default-to-slack"
  matchers {
    match_name  = "severity"
    match_value = ".*"
    match_type  = "regex"
  }
  channel_ids    = [signoz_notification_channel.slack.id]
  repeat_interval = "1h"
}

resource "signoz_alert" "error_rate" {
  alert      = "High Error Rate"
  alert_type = "LOGS_BASED_ALERT"
  severity   = "critical"
  condition  = jsonencode({ /* ... */ })
}
```

### Retention management

```hcl
resource "signoz_data_retention" "production" {
  metrics_ttl_duration_hrs = 2160  # 90 days
  logs_ttl_duration_hrs    = 720   # 30 days
  traces_ttl_duration_hrs  = 168   # 7 days
}
```

### Log pipeline chain

```hcl
resource "signoz_log_pipeline" "parse_json" {
  name     = "Parse JSON Logs"
  order_id = 1
  enabled  = true
  filter   = jsonencode({ op = "AND", items = [] })
  config   = jsonencode([{
    id = "json-parser", orderId = 1, enabled = true,
    name = "JSON Parser", type = "json_parser", parse_from = "body"
  }])
}

resource "signoz_log_pipeline" "extract_trace" {
  name     = "Extract Trace ID"
  order_id = 2
  enabled  = true
  filter   = jsonencode({ op = "AND", items = [] })
  config   = jsonencode([{
    id = "trace-extract", orderId = 1, enabled = true,
    name = "Trace ID", type = "trace_parser", parse_from = "attributes.trace_id"
  }])
}
```

---

## API Reference

| Resource | API Endpoint | Methods |
|----------|-------------|---------|
| `signoz_alert` | `/api/v1/rules` | GET, POST, PUT, DELETE |
| `signoz_dashboard` | `/api/v1/dashboards` | GET, POST, PUT, DELETE |
| `signoz_notification_channel` | `/api/v1/channels` | GET, POST, PUT, DELETE |
| `signoz_route_policy` | `/api/v1/route_policies` | GET, POST, PUT, DELETE |
| `signoz_downtime_schedule` | `/api/v1/downtime_schedules` | GET, POST, PUT, DELETE |
| `signoz_data_retention` | `/api/v1/settings/ttl` | GET, POST |
| `signoz_saved_view` | `/api/v1/explorer/views` | GET, POST, PUT, DELETE |
| `signoz_ingestion_key` | `/api/v2/gateway/ingestion_keys` | GET, POST, PATCH, DELETE |
| `signoz_log_pipeline` | `/api/v1/logs/pipelines` | GET, POST (batch-replace) |
