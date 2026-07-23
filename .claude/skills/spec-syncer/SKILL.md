---
name: spec-syncer
description: >-
  Re-run the full skaff codegen pipeline against the latest upstream SigNoz OpenAPI spec to catch up on spec drift, then open a PR with the regenerated code + docs. Use this when the task is "the upstream openapi spec changed", "regenerate the provider from the latest spec", "sync the generated code to  SigNoz/signoz main", "pick up the new spec", or a periodic spec-drift catch-up. This regenerates **every already-declared** resource — it does NOT add a new one (use resource-creator for that) and makes **no** `skaff.yml` edits.
---

# Sync generated code to the upstream OpenAPI spec

The upstream spec (`SigNoz/signoz` `main` → `docs/api/openapi.yml`) changes often. This provider tracks it: periodically regenerate everything from the current spec and land the drift as a **catch-up PR**. This is `resource-creator` minus Step 1 (declare) and Step 3 (register) — same pipeline, no `skaff.yml` change, all declared resources at once.

## Prerequisites

skaff runs through primus via the skaff-runner script — see the [skaff-runner](../skaff-runner/SKILL.md) and [primus-setter](../primus-setter/SKILL.md) skills. Set:

```sh
WT=<abs path to this repo>
MOD=github.com/SigNoz/terraform-provider-signoz
```

## Steps

### 1. Fetch the current spec + record its commit

Fetch the spec to `tmp/spec/openapi.yml` (see the **resource-creator** skill's Inputs for the `curl` and the not-plain-`main` merged-spec case).

```sh
mkdir -p tmp/spec

# capture the upstream commit for the PR body (reproducibility)
SPEC_SHA=$(gh api "repos/SigNoz/signoz/commits?path=docs/api/openapi.yml&per_page=1" --jq '.[0].sha')
echo "regenerating against SigNoz/signoz@$SPEC_SHA"
```

### 2. Branch

```sh
git switch -c chore/spec-sync-$(date +%Y%m%d)
```

### 3. Run the full pipeline

Run the **skaff-runner** script against the spec you fetched — it runs every step in the required order deterministically. Do **not** edit `skaff.yml`:

```sh
bash .claude/skills/skaff-runner/scripts/run-pipeline.sh tmp/spec/openapi.yml
```

The **skaff-runner** skill owns the `skipped 0` gate. The spec-sync twist: **if a schema that used to generate now `skipped`s**, the upstream spec drift broke a mapping — a schema gap, not a normal catch-up. **Stop, do not commit a partial regen.** Diagnose and surface it as per the known patterns in [.claude/docs/patterns.md](../../docs/patterns.md).

### 4. Dependencies (only if changed)

```sh
go mod tidy && go mod vendor   # ONLY if the regen changed go.mod
```

### 5. Gates

Run the Go gates — see the [`go-checks`](../../rules/go-checks.md) rule.

### 6. Regenerate docs

Run the **docs-writer** skill.

### 7. Review the diff — expected vs broken

A wholesale regen legitimately pulls in **unrelated** spec drift (e.g. a shared schema like `RenderError` gains/loses a field). That is the *point* — the provider tracks the current spec; commit it as catch-up.

Sanity-check the diff is a *catch-up*, not a *breakage*:
- ✅ field additions/removals on schemas, changed enum values, doc re-renders
- ❌ a `zz_generated_*` file **disappearing**, a resource's schema emptying out,
  or a new `skipped` — those mean a mapping broke (go to Step 3's stop rule)

### 8. Commit

One catch-up commit (regen output + regenerated docs together):

```sh
git commit -am "chore(codegen): sync generated code to upstream openapi spec

Regenerated against SigNoz/signoz@$SPEC_SHA."
```

### 9. Open the PR — confirm first

Opening a PR is outward-facing: **summarize the drift and confirm with the user before `gh pr create`.** Record `SigNoz/signoz@<sha>` and any notable drift in the PR body. Follow the [`pull-requests`](../../rules/pull-requests.md) rule for the template and attribution.

## Not this skill

- Adding a new `signoz_<name>` → **resource-creator** (declares in `skaff.yml`,
  registers in the provider).
- A dependency bump unrelated to the spec → a plain `build(deps)` PR.
