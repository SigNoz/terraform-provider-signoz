# Writing a guide (vN→vN+1)

The design and provider survey behind the SKILL's rule: **lead with `import {}` +
`terraform plan -generate-config-out`, never a bespoke JSON→HCL converter.** Shipped
precedent: `migrate-signoz-alert-to-signoz-rule` and the dashboard v1→v2 half of
`v0.0.x-to-v0.1.0` under `templates/guides/`.

## Recommendation: native config generation, not a converter

A "script that turns the exported JSON into a Terraform resource" is a hand-maintained
re-implementation of Terraform's own `-generate-config-out` — it would re-encode the
entire typed schema and re-sync on **every skaff regen**. Don't build it.

Lead every guide with native generation:

1. Write an `import {}` block (Terraform **1.5+**) pointing at the existing object's ID.
2. `terraform plan -generate-config-out=generated.tf` → Terraform writes the typed HCL
   from refreshed state.
3. `terraform apply` to execute the import; hand-clean `generated.tf`; `terraform plan`
   → "No changes".

This works because the v2 resources implement `ImportStatePassthroughID` **and** a
`Read` that does a `GET` + generated `Flatten*` convertor to populate the full typed
state — exactly the JSON→typed-model mapping a converter would need. Terraform then
serializes that state to HCL for free. The old CLI form `terraform import <addr> <id>`
populates **state but writes no config**, leaving the user to hand-write the whole
typed tree — switching to `import {}` + `-generate-config-out` is the single change
that makes migration easy.

## The workflow the guide should show

```hcl
# migrate.tf — requires Terraform 1.5+
import {
  to = signoz_rule.pod_cpu          # or signoz_dashboard.redis_overview
  id = "<existing-rule-or-dashboard-id>"
}
```

```sh
terraform plan -generate-config-out=generated.tf   # Terraform writes the typed HCL
terraform apply                                     # executes the import block
# hand-clean generated.tf: rename the label, trim computed-only attrs, eyeball
# deeply-nested blocks; move it into your real config; delete migrate.tf
terraform plan                                      # should report: No changes
```

For a resource-*type* change (alert→rule, or a v1→v2 dashboard swap) the old resource
must first leave state: `terraform state rm signoz_alert.<name>`.

**Honest caveats to document** (Datadog documents the identical ones):

- `-generate-config-out` emits verbose config incl. computed/optional attrs to trim.
- Deeply-nested blocks (panel/layout trees, composite queries) need eyeballing.
- But the output is **schema-correct by construction** — a hand-written converter can't
  guarantee that across regens.

### Optional bulk helper

A `scripts/` helper that lists all objects via the API and emits one `import {}` block
per object (with sanitized HCL labels) so users can bulk-generate. Mirror Datadog's
`scripts/cloud-cost-import-existing-resources/generate_import_config.sh`.

## Why this over the alternatives (provider survey)

| Approach | Who | Verdict |
| --- | --- | --- |
| `import {}` + `-generate-config-out` (TF 1.5+) | Datadog ships a full blueprint; Google documents a `terraform query` + `list {}` bulk variant (TF 1.14+, needs list resources) | ✅ **Recommended** — resources already import-ready; always matches the live schema. |
| JSON-passthrough attr (`*_json` string / `jsontypes.Normalized`) + diff normalizer | Datadog `*_json` resources; Google `google_monitoring_dashboard` | ❌ **Wrong direction** — the typed v2 resources exist to *delete* the JSON-string attrs; re-adding one re-introduces the drift they fix. |
| Bespoke JSON→HCL converter (`hclwrite` + cty walk) | grizzly / terraform-provider-grafana (not vendored) | ⚠️ **High-maintenance** — re-implements `-generate-config-out`; re-syncs every regen. |

None of the surveyed providers (AWS, Datadog, Google, PagerDuty) ships a bespoke
JSON→HCL writer — the ecosystem converged on `-generate-config-out`.

## If an offline converter is ever genuinely required

Only for authoring from a raw JSON file with **no live server**. Even then, don't
hand-roll an `hclwrite` walker — reuse the generated `Flatten*` convertors: a tiny
`cmd/` tool reads JSON → `json.Unmarshal` into the `apitypes` DTO → `Flatten*` → typed
model → serialize to HCL. The serialize step is the only new work — which is exactly
what `-generate-config-out` already does via cty. Simpler still: POST the JSON to a
scratch SigNoz and use the import flow above, so no new code is needed.
