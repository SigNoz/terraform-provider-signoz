---
page_title: "Migrating the fork's legacy resources to the upstream implementations"
subcategory: ""
description: |-
  The fork registered its own dashboard, notification_channel, route_policy and saved_view
  resources (legacy v1 APIs) and a broken downtime_schedule. The first four now use the
  upstream v2-API implementations; the broken downtime_schedule is removed outright, with
  `signoz_planned_maintenance` as its replacement. This guide maps every old schema to the new one.
---

# Migrating the fork's legacy resources

The fork's provider registered legacy, hand-written implementations under the type names
`signoz_dashboard`, `signoz_notification_channel`, `signoz_route_policy` and
`signoz_saved_view` (built against SigNoz v1 APIs), plus `signoz_downtime_schedule`, which
silently corrupted data (it wrote snake_case keys while the server reads camelCase). All
five are retired: the provider now registers upstream's generated resources for the four
overlapping names, and upstream's `signoz_planned_maintenance` covers downtime schedules.

Nothing is renamed — if you wrote `resource "signoz_dashboard" ...`, that block still
addresses a dashboard; only its **schema** changed. Do not upgrade and run `terraform
apply` directly: the new schemas address different server APIs, so plan would want to
destroy and recreate. Work through the flow below instead.

## Provider-level configuration changes (from the upstream rebase)

Two provider arguments changed meaning:

| Setting | Before (fork) | Now (upstream) |
|---|---|---|
| `endpoint` default | `http://localhost:3301` | `http://localhost:8080` |
| `http_timeout` | integer seconds, e.g. `35` | Go duration string, e.g. `"35s"` |

`http_timeout = 35` is now a plan-time error (`missing unit in duration`); write `"35s"`.

## The migration flow (Terraform 1.5+)

The new resources implement import plus a full typed `Read`, so Terraform can write the
new HCL for you — no hand-translation of `jsonencode(...)` blobs:

```hcl
# migrate.tf — temporary, delete once migration is done
import {
  to = signoz_dashboard.my_dashboard
  id = "<object-id>"
}
```

```sh
terraform state rm signoz_dashboard.my_dashboard     # 1. old schema leaves state
terraform plan -generate-config-out=generated.tf     # 2. Terraform writes typed HCL
terraform apply                                      # 3. executes the import
# 4. move generated.tf into your config, trim computed attrs, delete migrate.tf
terraform plan                                       # 5. must report: No changes
```

Caveats: the generated config is verbose (computed/optional attributes to trim) and
deeply-nested blocks (panel/layout trees, composite queries) need eyeballing — but it is
schema-correct by construction, unlike a hand-translated config.

## signoz_dashboard

`signoz_dashboard` is now the typed v2 (Perses-style) resource. Its full v1→v2 mapping —
including the note that **SigNoz migrates legacy v1 dashboards server-side** so they become
importable through the v2 API ([SigNoz/signoz#12177](https://github.com/SigNoz/signoz/issues/12177)) —
is documented in the upstream guide
[v0.0.x to v0.1.0](v0.0.x-to-v0.1.0.md); follow its `import {}` workflow above.

| Old (fork) | New (upstream) |
|---|---|
| `title` | `spec.display.name` |
| `description` | `spec.display.description` |
| `name` | `name` |
| `version` (e.g. `"v4"`) | `schema_version` (e.g. `"v6"`) |
| `widgets` (JSON string) | `spec.panels` (map keyed by panel id, each `{kind = "Panel", spec = {...}}`) |
| `layout` (JSON string) | `spec.layouts` (list of `{grid = {kind = "Grid", spec = {items = [...]}}}`) |
| `variables` (JSON string) | `spec.variables` (list) |
| `panel_map` (JSON string) | — (panels carry their own config now) |
| `tags` (list of strings) | `tags` (list of `{key, value}` objects) |
| `collapsable_rows_migrated`, `uploaded_grafana`, `source` | — (dropped) |

Before:

```hcl
resource "signoz_dashboard" "api_overview" {
  title       = "API Overview"
  name        = "api-overview"
  description = "API latency and error rate"
  version     = "v4"

  layout    = jsonencode([{ h = 6, i = "p-1", w = 12, x = 0, y = 0 }])
  widgets   = jsonencode([{ id = "p-1" }])
  variables = jsonencode({})
}
```

After (generated config is the reliable route; this shows the shape — see
`examples/resources/signoz_dashboard/resource.tf` for a full panel):

```hcl
resource "signoz_dashboard" "api_overview" {
  schema_version = "v6"
  name           = "api-overview"
  tags           = [{ key = "tag", value = "production" }]

  spec = {
    display = {
      name        = "API Overview"
      description = "API latency and error rate"
    }
    panels = {
      "api-latency" = {
        kind = "Panel"
        spec = {
          display = { name = "Latency" }
          plugin = {
            time_series_panel = {
              kind = "signoz/TimeSeriesPanel"
              spec = {}
            }
          }
          queries = []
        }
      }
    }
    layouts = [{
      grid = {
        kind = "Grid"
        spec = {
          items = [{
            x = 0, y = 0, width = 12, height = 6
            content = { ref = "#/spec/panels/api-latency" }
          }]
        }
      }
    }]
  }
}
```

Import id: the v2 dashboard UUID (`terraform import signoz_dashboard.x <uuid>` /
`import {}` block). If SigNoz has not migrated a particular v1 dashboard to the v2 store
yet, import cannot see it — `terraform state rm`, then re-create against the v2 API.

See [`docs/resources/dashboard.md`](../resources/dashboard.md).

## signoz_notification_channel

The fork wrote v1 `/api/v1/channels` receiver rows; the new resource manages
Alertmanager-style channels on the v2 API. Per-type config blocks collapse into one
`config` object keyed by channel type, each holding `kind` + `spec`.

| Old (fork) | New (upstream) |
|---|---|
| `name` | `name` (identifier) + `display_name` (human label; route policies reference this) |
| `type` (`slack`/`webhook`/`pagerduty`/`opsgenie`/`msteams`/`email`) | selects which `config.<type>` block you fill |
| `slack_configs = { api_url, channel, title, text }` | `config.slack = { kind = "slack", spec = { api_url, channel, ... } }` |
| `webhook_configs = { api_url, username, password }` | `config.webhook = { kind = "webhook", spec = { url, username, password, ... } }` |
| `pagerduty_configs = { routing_key, service_key, details }` | `config.pagerduty.spec = { ... }` |
| `opsgenie_configs` / `msteams_configs` / `email_configs` | `config.opsgenie` / `config.msteams` / `config.email` |

Before:

```hcl
resource "signoz_notification_channel" "slack" {
  name = "slack-alerts"
  type = "slack"

  slack_configs = {
    api_url = var.slack_webhook_url
    channel = "#alerts"
  }
}
```

After:

```hcl
resource "signoz_notification_channel" "slack" {
  name         = "slack-alerts"
  display_name = "Ops Slack"

  config = {
    slack = {
      kind = "slack"
      spec = {
        api_url       = var.slack_webhook_url
        channel       = "#alerts"
        send_resolved = true
      }
    }
  }
}
```

Import by the v2 channel id (`terraform import signoz_notification_channel.slack <id>`).
Legacy v1 channel ids do not resolve against the v2 API: try the import first, and where
the object is not visible, `terraform state rm` and re-create.

See [`docs/resources/notification_channel.md`](../resources/notification_channel.md).

## signoz_route_policy

The fork mapped Prometheus route fields one-by-one; upstream exposes the v2 route as
`name` + a filter `expression` + `channels` referenced by **channel display name** (not id).

| Old (fork) | New (upstream) |
|---|---|
| `route_name` | `name` (required) |
| `matchers { match_name, match_value, match_type }` blocks | `expression` string, e.g. `service == "payments" && severity == "critical"` |
| `channel_ids = [<id>]` | `channels = [<channel display_name>]` |
| `group_wait` / `group_interval` / `repeat_interval` | — (server-managed) |
| `enabled` / `continue_matching` | — (dropped) |
| `description` | `description` |
| — | `kind` (`"rule"` or `"policy"`; optional/computed) |
| — | `tags` (list of strings) |

Before:

```hcl
resource "signoz_route_policy" "critical" {
  route_name = "critical-alerts"

  matchers {
    match_name  = "severity"
    match_value = "critical"
    match_type  = "equals"
  }

  channel_ids = [signoz_notification_channel.slack.id]
  group_wait  = "30s"
}
```

After:

```hcl
resource "signoz_route_policy" "critical" {
  name       = "critical-alerts"
  kind       = "policy"
  expression = "severity == \"critical\""
  channels   = [signoz_notification_channel.slack.display_name]
}
```

Import by the v2 route id. As with channels, v1 route ids do not resolve on the v2 API —
import where possible, otherwise `terraform state rm` + re-create.

See [`docs/resources/route_policy.md`](../resources/route_policy.md).

## signoz_saved_view

The fork stored the whole explorer query as an `explorer_data` JSON string; upstream models
the view as a typed `spec` with structured `queries`.

| Old (fork) | New (upstream) |
|---|---|
| `name` | `name` |
| `source_page` (`logs`/`traces`/`metrics`) | `source` |
| `explorer_data` (JSON string) | `spec` (typed: `request_type`, `panel_type`, `queries`, `display_name`, `selected_fields`) |
| `category` | — (implied by `spec` query shape / panel_type) |
| `extra_data` | — (dropped) |
| `tags` | — (dropped) |
| `uuid` | `id` (computed) |
| — | `schema_version` (required) |

Before:

```hcl
resource "signoz_saved_view" "error_logs" {
  name        = "Error Logs - Production"
  source_page = "logs"
  tags        = ["production"]

  explorer_data = jsonencode({
    query = { queryType = "builder" }
  })
}
```

After (see `examples/resources/signoz_saved_view/resource.tf` for the full query):

```hcl
resource "signoz_saved_view" "error_logs" {
  name           = "api-service-errors"
  schema_version = "v2"
  source         = "logs"

  spec = {
    display_name = "Error Logs - Production"
    panel_type   = "list"
    request_type = "raw"

    queries = [{
      builder_query = {
        type = "builder_query"
        spec = {
          logs = {
            name   = "A"
            signal = "logs"
            filter = { expression = "severity_text = 'ERROR'" }
          }
        }
      }
    }]
  }
}
```

Import by the view id. Views created through the v1 explorer API may not exist on the v2
API: `terraform state rm` + re-create wherever the import cannot find them.

See [`docs/resources/saved_view.md`](../resources/saved_view.md).

## signoz_downtime_schedule → signoz_planned_maintenance

`signoz_downtime_schedule` is removed outright: it sent snake_case keys (`start_time`)
while the server expects camelCase (`startTime`), so windows were created with silently
dropped fields. The upstream `signoz_planned_maintenance` resource addresses the same
feature correctly and is already registered.

| Old (fork) | New (upstream) |
|---|---|
| `name` | `name` (required) |
| `description` | `description` |
| `schedule { start_time, end_time, timezone }` block | `schedule = { start_time, end_time, timezone }` (object) |
| `recurrence { type, days_of_week, duration }` | `recurrence = { repeat_type, repeat_on, duration }` (`repeat_on` = list of strings) |
| `alert_ids` | `alert_ids` |
| `all_alerts = true` | omit `alert_ids` and leave `scope` empty — a maintenance with an empty `alert_ids` set applies to all alerts |
| `id` | `id` (computed), `kind` (computed) |

`scope` is not a keyword or enum: the server stores it as an
[expr-lang](https://expr-lang.org/docs/language-definition) boolean expression evaluated
against each alert's labels, and rejects unparseable expressions when you write the
maintenance. Leave `alert_ids` empty to cover every alert, then use `scope` to narrow
which of them are actually suppressed, e.g. `scope = "severity == \"critical\""`.
See SigNoz's own documentation section,
[Scoping with label expressions](https://signoz.io/docs/alerts-management/planned-maintenance/#scoping-with-label-expressions).

Before:

```hcl
resource "signoz_downtime_schedule" "weekly" {
  name       = "Weekly Maintenance"
  all_alerts = true

  schedule {
    start_time = "2026-01-15T02:00:00Z"
    end_time   = "2026-01-15T04:00:00Z"
    timezone   = "America/New_York"

    recurrence {
      type         = "weekly"
      days_of_week = [0, 6]
      duration     = "2h"
    }
  }
}
```

After:

```hcl
resource "signoz_planned_maintenance" "weekly" {
  name = "Weekly Maintenance"

  # No alert_ids and no scope: the window suppresses every alert.
  schedule = {
    start_time = "2026-01-15T02:00:00Z"
    end_time   = "2026-01-15T04:00:00Z"
    timezone   = "America/New_York"

    recurrence = {
      repeat_type = "weekly"
      repeat_on   = ["0", "6"]
      duration    = "2h"
    }
  }
}
```

Because the legacy resource wrote corrupt data, existing `signoz_downtime_schedule` state
may not match a usable server object. `terraform state rm` each one, check what actually
landed in SigNoz (delete leftovers in the UI), then manage windows with
`signoz_planned_maintenance` — importing by id only where the maintenance window is intact.

See [`docs/resources/planned_maintenance.md`](../resources/planned_maintenance.md).

## What about `signoz_alert`?

Unchanged. The fork's `signoz_alert` (v1 rules API) stays registered alongside upstream's
`signoz_rule` (v2 rules API); migrating `alert → rule` is covered by
[`migrate-signoz-alert-to-signoz-rule.md`](migrate-signoz-alert-to-signoz-rule.md).
