---
name: integration-tester
description: >-
  Run the end-to-end integration suite (spins up a real SigNoz via foundry, drives Terraform against a locally-built provider, asserts no drift), test a resource change against a live server, add or debug a `tests/testdata/` scenario, or investigate a `terraform plan` that won't converge. Use this whenever the task is "run the integration tests", "test this against a real SigNoz", "add a testdata scenario", "why is there drift after apply", or verifying a create/update round-trips cleanly. The suite lives under `tests/` (Python + pytest + uv).
---

@tests/README.md

## Standing rules

- **Pause after create/update verification.** When walking a resource through a
  flow by hand, stop once create/update + no-drift is confirmed — don't auto-chain
  into `terraform destroy`; let the user decide.
- Test-writing conventions (fixtures, skip-at-collection, config via `--flags` not
  env, no `_helper()`) are the **pytest** rule — it auto-loads on any `.py` file.

## Writing a testdata scenario

The README covers the *mechanics* — layout (`resources/signoz_<name>/NN/`, a single
`01-<name>.{tf,tf.json}` base + optional ordered `NN-jsonpatch.json` edits) and the
create → no-drift → [edit → drift → converge] → destroy lifecycle the runner drives.
This is *what* each scenario should prove.

**One scenario isolates one thing that must round-trip cleanly** (`<name>` says
what). Aim for this coverage per resource:

- **Baseline** — one representative full config (`00`): create → no-drift → destroy.
- **Edit path** — a `.tf.json` base + `NN-jsonpatch.json` (RFC 6902) edits, proving
  in-place updates converge (re-plan drifts → apply → re-plan clean). Use `.tf.json`
  **whenever you'll patch** — a JSON Patch needs a JSON target; a patch-free scenario
  can stay plain `.tf`.
- **Each variant the resource supports** — one scenario apiece: query type
  (builder / clickhouse / promql / anomaly), operator (below / absent), rule type,
  panel / variable kind, …
- **Drift-prone edge values** — the shapes the server tends to drop, enrich, or
  normalize (see "When a round-trip won't converge"): empty enums; empty / zero /
  `null` / omitempty fields; every state of an optional sub-object (renotify off,
  `group_by` empty, `use_policy` false, `alert_states` empty); and string↔number
  coercions (e.g. `step_interval` as a string). This is where drift actually shows up.
- **Real-world configs** — for a complex resource (dashboard), a few real exported
  configs converted to typed HCL, to catch shapes synthetic scenarios miss.

Keep each scenario **minimal-but-valid** — the smallest config that isolates the
thing under test. A scenario that references another resource (e.g. a channel) must
have it seeded; the suite seeds `slack` / `pagerduty`.

## When a round-trip won't converge

Persistent post-apply drift is usually a **server-side round-trip** issue, not a
provider bug — the API drops or enriches fields so a create→read isn't stable (e.g.
`notificationSettings.renotify` dropped, `condition` enriched with unset optionals,
durations normalized `5m`→`5m0s`). That's a fix for the SigNoz backend, not the
provider.
