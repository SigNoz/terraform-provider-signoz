---
name: skaff-runner
description: >-
  Run the ENTIRE skaff codegen pipeline deterministically from an OpenAPI spec, via a bundled script — every declared resource, all steps in the required order (types → schemas → apitypes → client → flex → services → convertors last). Use whenever you need to (re)generate the provider's `internal/**` code from a spec. Called by the resource-creator and spec-syncer skills; run it directly when you just need a clean regen. It does not edit `skaff.yml`, register resources, regenerate docs, or open a PR.
---

# skaff pipeline runner

`scripts/run-pipeline.sh` runs the whole skaff pipeline in one deterministic pass, so the order and flags are never retyped by hand. It regenerates **every** resource declared in `skaff.yml` from the spec you pass.

```sh
bash .claude/skills/skaff-runner/scripts/run-pipeline.sh <openapi-spec-path>
```

- **`<openapi-spec-path>`** — the spec to generate from (e.g. one fetched to `tmp/spec/openapi.yml`). Required.
- **`DRY_RUN=1`** — print each skaff invocation instead of running it (preview).
- **`SKAFF_VERSION=vX.Y.Z`** — pin the skaff release primus downloads.

What it guarantees (so callers don't have to remember):
- steps run in order, **`convertors` last** with `--callers-dir` (reachability prune), and `--convtypes-import` on **both** `flex` and `convertors`;
- repo root and Go module path are derived automatically (`git`, `go.mod`);
- skaff runs through **primus** — needs `PRIMUS_HOME` (or a `./.primus` clone); if primus isn't set up it errors and points to the **primus-setter** skill.

## After it runs

- `flex` and `services` must report **`skipped 0`**. A skip = a resource that
  isn't Pattern A (field-for-field) or a schema gap — stop and investigate: match
  the shape in the pattern catalog (imported under **Reference** below).
- The pipeline touches every declared resource, so `git status` shows byte-
  identical re-renders of the others alongside your change — expected.
- Then build / lint (`go-checks` rule) and, if schemas changed, regenerate docs
  (`docs-writer` skill).

## Reference

The pattern catalog — the doc you need to handle a skip: it maps each shape to its
fix and links onward to the deeper design docs in `.claude/docs/`.

@.claude/docs/patterns.md

## Used by

- **resource-creator** — Step 2 (generate) calls this script.
- **spec-syncer** — Step 3 (regenerate all from the latest spec) calls this script.
