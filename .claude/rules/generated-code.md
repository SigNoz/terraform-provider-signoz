---
paths:
  - "internal/**"
---

# Generated code under internal/

Most of `internal/` is **generated** by the `SigNoz/skaff` codegen pipeline from the SigNoz OpenAPI spec + `skaff.yml`. Do not hand-edit generated files — changes are lost on the next regen. To change generated code, fix the input (the SigNoz schema or `skaff.yml`) and re-run the pipeline. See the **resource-creator** skill for the full runbook. Generated files start with  `zz_generated_*` prefix.

**Hand-written — safe to edit, preserved across regen:**

- `internal/apiclients/{wrapped_client,doer,error}.go` — retry/auth/error HTTP wrapper
- `internal/convtypes/*.go` — generic conversion helpers (`StringPointer`, `TimeFromString`, …)
- `internal/provider/provider.go` — framework provider entry; registers each resource

When editing hand-written Go here, follow the `go-comments` and `go-newlines` rules.
