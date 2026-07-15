---
page_title: "Migrating from signoz_alert to signoz_rule"
subcategory: ""
description: |-
  How to migrate existing signoz_alert resources to the typed signoz_rule resource on the SigNoz v2 rules API.
---

# Migrating from `signoz_alert` to `signoz_rule`

`signoz_rule` is the recommended resource for alert rules. It is built on the SigNoz **v2 rules API** (`/api/v2/rules`) and is generated from the SigNoz OpenAPI spec, so every field is a **typed, first-class attribute** that Terraform validates at plan time.

`signoz_alert` (the older resource on `api/v1/rules`) stores most of its important attributes, as opaque `jsonencode(...)` strings. Terraform cannot plan, validate, or diff the contents of those blobs, so a typo inside a query or threshold is only caught at apply time, and any server-side normalization of the JSON shows up as permanent plan noise.

`signoz_rule` removes both problems.

-> **`signoz_alert` is not being removed yet.** Both resources ship side by side and existing `signoz_alert` configurations keep working against `api/v1/rules`. A formal deprecation and removal will get its own release note.

## `signoz_rule` is v2alpha1-only

`signoz_rule` supports a single schema version: **`schema_version =
"v2alpha1"`**. It always sends and expects the v2alpha1 payload, and every
example and generated field in this guide assumes it. Set it explicitly on every
`signoz_rule`:

```terraform
resource "signoz_rule" "example" {
  # ...
  schema_version = "v2alpha1"
}
```

Both resources describe the *same* v2alpha1 rule payload — `signoz_alert` just carries attributes as JSON strings, while `signoz_rule` carries them as typed attributes. Migration is therefore mostly a matter of turning those JSON blobs into typed HCL: camelCase keys become the provider's snake_case, and a couple of structures gain a wrapper. You do **not** have to do that translation by hand — the recommended path below has Terraform generate it for you.

## Before you begin

- **The rule on the server must already be at `v2alpha1`.** `signoz_rule` only speaks v2alpha1 — it reads and writes the rule through `/api/v2/rules` as v2alpha1. If the rule is still stored under the legacy **v1** schema (the default `schema_version = "v1"` on `signoz_alert`), `signoz_rule` cannot import, read, or manage it, and the steps below will not work. Convert it first — see [Convert a v1 rule to v2alpha1 first](#convert-a-v1-rule-to-v2alpha1-first).
- **Terraform 1.5 or newer**, for the recommended import-block workflow (`import {}` + `-generate-config-out`).
- The rule already exists on the server. The v1 and v2 APIs are two views over one rule store, so a rule created through `signoz_alert` can be adopted by `signoz_rule` using its ID — no rule is created, deleted, or interrupted during migration.
- **Record each rule's ID first.** You need it to import, and it leaves state when you remove the old resource:

```sh
terraform state show signoz_alert.<name>   # copy the `id` value
```

### Convert a v1 rule to v2alpha1 first

`signoz_rule` manages only v2alpha1 rules. If the rule on the server is still on the v1 schema, Terraform cannot do anything with it as a `signoz_rule` — both the import and the `-generate-config-out` step will fail. Upgrade the rule to v2alpha1 **before** migrating, while it is still a `signoz_alert`:

1. On the existing `signoz_alert`, set `schema_version = "v2alpha1"` and express `condition` / `evaluation` / `notification_settings` in the v2alpha1 shape (the same body `signoz_rule` uses — see the [attribute mapping](#alternative-rewrite-the-config-by-hand) below).
2. Run `terraform apply`. This rewrites the underlying rule to v2alpha1 on the server; it stays a `signoz_alert` for now.
3. Confirm the rule now reports `schema_version = "v2alpha1"` — check `terraform state show signoz_alert.<name>`, or open the alert in the SigNoz UI.

Rules created directly on the v2 API, or that already report `schema_version = "v2alpha1"`, need no conversion — skip straight to the migration steps below.

-> **You may not need to do this by hand.** An upcoming SigNoz release will migrate v1 rules to v2alpha1 automatically. If you would rather not convert them yourself, wait for that release; once your rules report `schema_version = "v2alpha1"`, follow the migration steps below.

## Recommended: let Terraform generate the typed config

Because `signoz_rule` already implements import plus a full typed read (a `GET` on the v2 API mapped into the typed model), Terraform can write the v2alpha1 HCL for you with `-generate-config-out`. This is the easiest and least error-prone path - the generated config is schema-correct by construction and always matches the live rule.

For each alert (`<name>` = the resource label, `<rule-id>` = the ID you
recorded):

1. **Remove the old resource from state.** This is a resource-*type* change, not
   an in-place upgrade, so the `signoz_alert` entry has to leave state first:

   ```sh
   terraform state rm signoz_alert.<name>
   ```

2. **Add an `import` block** pointing the new resource at the existing rule ID:

   ```terraform
   # migrate.tf — delete once migration is done
   import {
     to = signoz_rule.<name>
     id = "<rule-id>"
   }
   ```

3. **Generate the typed config.** Terraform reads the rule through the v2 API and
   writes the full typed `signoz_rule` block:

   ```sh
   terraform plan -generate-config-out=generated.tf
   ```

4. **Execute the import:**

   ```sh
   terraform apply
   ```

5. **Clean up.** Move the block from `generated.tf` into your real configuration,
   then:
   - delete the old `signoz_alert` block from your config,
   - delete `migrate.tf` and `generated.tf`,
   - trim the generated block if you like. `-generate-config-out` emits every
     optional and computed attribute, so it is verbose; deeply nested blocks (the
     composite-query tree) are worth eyeballing.

6. **Confirm no drift:**

   ```sh
   terraform plan   # should report: No changes
   ```

Repeat per rule. Everything you have not migrated keeps running under
`signoz_alert`.

-> The generated config is emitted directly from what the v2 API returns for the
rule, so it is always valid `v2alpha1`. If step 6 shows a diff, treat it as real
drift (or a bug to report) rather than editing blindly.

## Alternative: rewrite the config by hand

If you would rather translate the config yourself — or you are authoring a new
`signoz_rule` from an old `signoz_alert` you have on file — the mapping is
mechanical. Keep the `import` step (steps 1-2 and 4 above): the rule still needs
to be adopted by ID; only the config for step 3 is written by hand instead of
generated.

### Top-level attributes

| `signoz_alert` (v1) | `signoz_rule` (v2, v2alpha1) |
| --- | --- |
| `alert`, `alert_type`, `rule_type` | same attributes, all **required** |
| `schema_version` | `schema_version` — **required**; must be `"v2alpha1"` |
| `description`, `disabled` | same attributes (optional) |
| `summary` (flat string) | `annotations.summary` (`annotations` is a map) |
| `severity` (flat string) | `labels.severity` (`labels` is a map) |
| `condition` (`jsonencode(...)` string) | `condition` (typed object, **required**) |
| `evaluation` (`jsonencode(...)` string) | `evaluation` (typed object, **required**) |
| `eval_window`, `frequency` (flat, older alert shape) | `evaluation.rolling.spec.eval_window` / `.frequency` |
| `notification_settings` | `notification_settings` — same shape, but now **required** (use `{}` if you have none) |
| `labels` | `labels` — same (map) |
| `preferred_channels` | **removed** — route per threshold via `condition.thresholds.basic.spec[].channels`, plus `notification_settings` |
| `version` | **removed** — no `signoz_rule` equivalent; drop it |
| `source` | **removed** — no `signoz_rule` equivalent; drop it |
| `broadcast_to_all` | **removed** — no `signoz_rule` equivalent; drop it |

Eight `signoz_alert` fields have no top-level home on `signoz_rule`. Four **move** into typed attributes and are still expressible — `summary` → `annotations.summary`, `severity` → `labels.severity`, and `eval_window` / `frequency` → `evaluation.rolling.spec`. The other four — `preferred_channels`, `version`, `source`, and `broadcast_to_all` — are **removed outright** with no equivalent; drop them from the config.

### Inside `condition`

`condition` changes in two ways: its keys are renamed, and a few structures are
re-nested. The full before/after is in the [worked example](#worked-example)
below; this is the reference.

**1. Keys become snake_case.** Every key in the JSON body is renamed mechanically:

| `signoz_alert` (JSON) | `signoz_rule` (HCL) |
| --- | --- |
| `compositeQuery` | `composite_query` |
| `queryType` | `query_type` |
| `panelType` | `panel_type` |
| `stepInterval` | `step_interval` |
| `metricName` | `metric_name` |
| `spaceAggregation` | `space_aggregation` |
| `timeAggregation` | `time_aggregation` |
| `fieldContext` | `field_context` |
| `fieldDataType` | `field_data_type` |
| `selectedQueryName` | `selected_query_name` |
| `matchType` | `match_type` |
| `targetUnit` | `target_unit` |
| `alertOnAbsent` | `alert_on_absent` |
| `absentFor` | `absent_for` |

**2. Structures are re-nested.** On top of the renames:

- **Each query becomes a typed variant.** A `compositeQuery.queries[]` entry
  (`{ type = "...", spec = {...} }`) is wrapped in a block named for its kind —
  `builder_query`, `builder_formula`, `promql`, `clickhouse_sql`, or
  `builder_trace_operator` — each carrying `type` and `spec`.
- **A builder query's `spec` gains a signal sub-object.** In a `builder_query`,
  the signal-specific fields (`name`, `aggregations`, `filter`, `group_by`,
  `legend`, `step_interval`, `disabled`, …) move under a `metrics`, `logs`, or
  `traces` block chosen by `signal`. Formulas, PromQL, and ClickHouse queries
  keep a flat `spec`.
- **Thresholds gain a kind wrapper.** `thresholds = { kind = "basic", spec = [...] }`
  becomes `thresholds = { basic = { kind = "basic", spec = [...] } }`.

### Inside `evaluation`

The evaluation JSON gains a wrapper naming its kind:

```terraform
# signoz_alert
evaluation = jsonencode({
  kind = "rolling"
  spec = { evalWindow = "15m", frequency = "1m" }
})

# signoz_rule
evaluation = {
  rolling = {
    kind = "rolling"
    spec = {
      eval_window = "15m"
      frequency   = "1m"
    }
  }
}
```

A cumulative evaluation wraps under `cumulative` instead of `rolling`.

## Migrating many rules at once

For a large number of alerts, generate the `import` blocks programmatically: list your rules through the SigNoz API, emit one `import {}` block per rule (with a sanitized HCL label), and run a single `terraform plan -generate-config-out=generated.tf` over all of them. The steps are otherwise identical to the single-rule flow above.
