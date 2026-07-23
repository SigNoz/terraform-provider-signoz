# Audit Fields

Whether the provider should keep `created_at` / `updated_at` / `created_by` / `updated_by` in resource state or other server side generated fields.

## Problem

SigNoz API responses carry server-set audit fields. If they land in resource state they cause **perpetual plan drift**: `updated_at` changes on every API touch, and auditor identity (`*_by`) is meaningless for IaC (the principal running Terraform is the author by definition).

## What we decided, and why

**Drop all four from both resource *and* data-source schemas.** This sits between "keep with care" and "strip aggressively". Timestamps could be kept with plan modifiers, but they carry maintenance cost and no strong output use case here, so the current decision is to drop them everywhere. If audit data is ever genuinely needed, the natural place to reintroduce it is a **data source** (read-only, so drift/plan-noise don't apply) — mirroring PagerDuty's deprecation move — but that is a future change, not today's default.

Keep `id` (framework + import) and any server-derived field genuinely useful as an output reference; give stable-post-create fields `UseStateForUnknown`.

## How to do it

Via `skaff.yml`'s `schema.ignores`, so the decision is codegen-driven and lives in one place. Also drop the response envelope (`data`, `status`):

```yaml
<resource>:
  schema:
    ignores: [data, status, createdAt, createdBy, updatedAt, updatedBy]
```

This is *why* every resource block in `skaff.yml` ignores those six keys. The data source's `schema.ignores` drops the same six keys, so audit fields are absent there too.


## External References

Surveyed AWS, PagerDuty, Google (13 resources):

- **`created_by` / `updated_by` — universally dropped.** Zero of the 13 expose auditor identity; it's treated as out of scope for IaC.
- **AWS** — keeps timestamps, carefully: `created_at` is `Computed` + `UseStateForUnknown` (immutable, no churn); `updated_at` is bare `Computed` (correctly shows `(known after apply)`). The most permissive, highest-maintenance pattern.
- **PagerDuty** — drops them on most resources, and has a documented **deprecation precedent**: `last_incident_timestamp`/`status` were removed with the comment *"caused persistent drift in plan output. Use `data.pagerduty_service` if you need this value."*.
- **Google** — keeps immutable creation timestamps (no churn), drops identity.
