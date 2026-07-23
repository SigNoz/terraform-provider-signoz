---
name: resource-creator
description: >-
  End-to-end workflow for adding a new Terraform resource + data source to terraform-provider-signoz via the skaff OpenAPI codegen pipeline. Use this whenever the task is to "add", "create", "generate", "scaffold", or "wire up" a new `signoz_<name>` resource/data source from a SigNoz API endpoint — even if the user just names the API (e.g. "dashboard v2 APIs", "downtime schedules", "service accounts") without saying "skaff" or "codegen". It covers declaring the resource in skaff.yml, running the skaff generators in the right order with the right flags, registering the resource in the live provider, generating examples + docs, the build/lint gates, and the two-commit split. Reach for it before hand-writing any resource — most resources are field-for-field and fully generated; a skip is a schema gap to fix upstream, never a convertor to hand-write.
---

# Adding a resource to terraform-provider-signoz

This provider is **generated**, not hand-written. A tool called [`SigNoz/skaff`](https://github.com/SigNoz/skaff) reads the SigNoz OpenAPI spec plus a per-resource config (`skaff.yml`) and emits the `zz_generated_*` files across `internal/`. Adding a resource means *declaring* it and *running the pipeline* — you almost never write resource logic by hand.

Many resources already shipped this way. Follow their shape; don't invent a new one.

## Mental model: the layers

@.claude/docs/architecture.md

## Inputs (where things live)

Two inputs you provide; skaff itself comes from primus:

- **Spec:** the canonical SigNoz OpenAPI spec (`SigNoz/signoz`, `main`, `docs/api/openapi.yml`). Fetch a fresh copy — never a stale in-repo copy:

```sh
mkdir -p tmp/spec
curl -fsSL https://raw.githubusercontent.com/SigNoz/signoz/main/docs/api/openapi.yml -o tmp/spec/openapi.yml
```

If you need a spec that isn't plain `main` — e.g. an open PR's contract merged in — hand-produce the merged spec into that same path, `tmp/spec/openapi.yml`.

- **Config:** `skaff.yml` at this repo root (you edit this in Step 1).

`skaff` runs through **primus** — the skaff-runner script (Step 2) invokes it, so you never fetch or build the binary yourself. If primus isn't set up, run the **primus-setter** skill first.

## Step 1 — Declare in `skaff.yml`

Add one block under `resources:` (create/read/update/delete: each a `path` + `method`) and one under `data_sources:` (read only). Copy the nearest existing block. The path template uses `{id}`. Example (collection):

```yaml
resources:
  <name>:
    create: { path: /api/v1/<collection>, method: POST }
    read:   { path: /api/v1/<collection>/{id}, method: GET }
    update: { path: /api/v1/<collection>/{id}, method: PUT }
    delete: { path: /api/v1/<collection>/{id}, method: DELETE }
    schema:
      ignores:
        - data        # response envelope
        - status      # also drops a *meaningful* status field if the resource has one (accepted gap)
        - createdAt
        - createdBy
        - updatedAt
        - updatedBy
data_sources:
  <name>:
    read: { path: /api/v1/<collection>/{id}, method: GET }
    schema:
      ignores: [ data, status, createdAt, createdBy, updatedAt, updatedBy ]
```

**`schema.ignores` rules:**
- Drops the response envelope + server audit fields — see [`audit-fields.md`](../../docs/audit-fields.md) for why.
- To hide a nested attribute, add its **OpenAPI property name** to `ignores`. This keeps the model scalar and stops `skaff types` emitting custom types for it (service_account ignores `serviceAccountRoles`; user ignores `userRoles` and `frontendBaseUrl`).
- Ignoring everything nested → an all-scalar resource (route_policy, service_account-after-ignore) needs **no** customtypes and no per-customtype helpers — it still gets a `flex` entry point.

## Step 2 — Run the pipeline

Run the **skaff-runner** script — it runs the whole pipeline deterministically, through primus. Pass the spec you fetched:

```sh
bash .claude/skills/skaff-runner/scripts/run-pipeline.sh tmp/spec/openapi.yml
```

Preview the exact skaff commands first with `DRY_RUN=1 …`. See the [skaff-runner skill](../skaff-runner/SKILL.md) for the flags and prerequisites.

The pipeline regenerates **every** declared resource, not just the new one. That's expected — `git status` will show your new files plus byte-identical re-renders of the others.

### Watch the per-step output

- **`skaff types`** prints how many customtypes are needed. A nested object (planned_maintenance's `schedule` → Schedule/Recurrence) brings in `internal/customtypes/` + per-customtype convertors; an all-scalar resource needs none.
- **`flex` and `services`** report a `skipped` count. A skip means the resource can't be generated field-for-field — treat it as a **schema gap to fix upstream**, not a convertor to hand-write (see "When skaff skips a schema" below). skaff-runner owns the `skipped 0` gate; see the [skaff-runner skill](../skaff-runner/SKILL.md).

## Step 3 — Register in the live provider

Edit `internal/provider/provider.go`:
- add `services.New<Name>Resource` to `Resources()`
- add `services.New<Name>DataSource` to `DataSources()`

Keep the `services.*` entries grouped (alphabetical), before any hand-written `signoz.*` entries. The generated services set `TypeName = ProviderTypeName + "_<name>"`, so the Terraform type is `signoz_<name>`. Registration must happen before docs (Step 4) — tfplugindocs introspects the live provider schema.

## Step 4 — Examples + docs

Run the **docs-writer** skill.

## Step 5 — Gates

Run `go mod tidy` + `go mod vendor` **only if a new dependency appeared** (the first nested-object resource pulls in `terraform-plugin-go`). Then run the Go gates — see the [`go-checks`](../../rules/go-checks.md) rule.

## Step 6 — Integration tests

## Step 7 — Commit (three-commit split)

On a `feature/<name>` branch, split the work in two commits:
1. **regen output:** `skaff.yml` + the `internal/**/zz_generated_*` files
2. **wiring + docs:** `internal/provider/provider.go`, `examples/.../signoz_<name>/**`, `docs/**/<name>.md`
3. **integration tests:** `tests/`

Then open a PR following the **pull-requests** rule (`.claude/rules/pull-requests.md`) — it owns the PR template and body format.

## When skaff skips a schema — fix the schema, not skaff

Run `skaff types`; for each skipped schema it prints the name + failing OAS path + reason. Log it, fix it upstream in the SigNoz schema, regenerate, and repeat until `skipped 0`.

The codegen references live in `.claude/docs/` — the catalog of known patterns and their fixes is in [`patterns.md`](../../docs/patterns.md). Read it before starting a complex resource, and recognize the shape there before hand-fixing.
