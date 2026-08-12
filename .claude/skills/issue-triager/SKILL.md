---
name: issue-triager
description: >-
  Triage a reported bug against terraform-provider-signoz: reproduce it, decide whether it is real and where the defect actually lives, add the coverage that proves it, document the shape in the pattern catalog, and hand over a suggested upstream fix. Use whenever the task is to "triage" — given a GitHub issue link (e.g. "triage SigNoz/terraform-provider-signoz#154") or just a prose description of a failure ("a user says an invalid dashboard passes plan then fails at apply") — and whenever a report needs confirming before anyone writes a fix. It ends at a suggested fix plus a draft PR carrying the reproduction; it does NOT implement the fix, comment on the public issue, or close it unless the user asks.
---

# Triage a reported bug

A report tells you what someone saw, not what is wrong. Both triages this repo has run ended somewhere other than where the reporter pointed: issue #154 blamed the provider for never sending `signal` (it always did — the real defect was a missing plan-time validator), and issue #155 read as a type-width bug but reproduced as a syntax mistake against a field whose docs never explain its real requirement. **Start from the error string, not the reporter's diagnosis.**

The provider is generated, so triage has one question the reporter can't answer: does the defect live in this repo at all? Read the [`generated-code`](../../rules/generated-code.md) rule before you conclude anything.

## Steps

### 1. Read the report and load the conventions

Get the issue body, every comment, and the reporter's exact config and diagnostic. With no link, work from the prose description the same way — a repro is a repro.

Write down two things separately: **the error string**, which is evidence, and **the mechanism the reporter proposes**, which is a hypothesis and has been wrong both times.

### 2. Reproduce it exactly as filed

Build the provider and drive the reporter's config through `terraform plan` / `apply` against a `dev_overrides` build. Reproduce the diagnostic **verbatim** before theorising.

Three outcomes, all of them useful:

- **Reproduces as filed** — go to Step 3.
- **Reproduces, different mechanism** — the report is real, the diagnosis isn't. Find the actual cause, and expect to correct the reporter in Step 8.
- **Doesn't reproduce** — find what *does* produce their exact diagnostic. When the answer is "the config was wrong", the defect is usually in the **user-facing surface**: check whether the docs render the field's real requirement, whether any example shows the working form, and whether the suite sets the field at all. A field that is undocumented, unexampled and untested is the finding.

### 3. Locate the defect — which repo owns it

Read the generated schema for the affected resource (`internal/schemas/zz_generated_<name>_resource.go`) and compare it against the spec. Then place the defect:

| Symptom | Owner |
|---|---|
| A spec constraint is missing from the plan-time schema | **skaff** — a post-processing traversal is incomplete |
| The spec itself is wrong, or a field needs `required` | **SigNoz** — `docs/api/openapi.yml` comes from Go struct tags |
| The server won't round-trip what it accepted | **SigNoz** backend |
| Docs, examples, or missing coverage | **this repo** |

**A generated-output bug is never fixed in the provider.** A guard committed here is wiped by the next regen. The two highest-yield checks:

- **The same constraint guarded on one resource and absent on another** — suspect the traversal, not the spec. Every skaff post-processing pass hand-rolls its own attribute walk and each is incomplete differently; a `map_nested` attribute was an opaque wall to the oneOf validators until [skaff#14](https://github.com/SigNoz/skaff/pull/14). See [`patterns.md`](../../docs/patterns.md), "Spec constraint that never reaches the schema".
- **Passes plan, fails at apply** — the signature of a constraint that exists in the spec and never became a schema constraint.

Verify every hypothesis you read out of source by running it. Reading a validation path in the SigNoz source is not a reproduction.

### 4. Add the coverage that proves it, in a new PR

Open a **draft PR** carrying the reproduction and nothing else. The fix lands later, after the upstream change ships.

- A **`tests/testdata/resources/signoz_<name>/NN/`** scenario at the next free two-digit number — see the [integration-tester](../integration-tester/SKILL.md) skill.
- Or an **`examples/resources/signoz_<name>/*.tf`** file. Those are themselves live integration tests: `test_examples.py` runs apply → plan-no-drift → destroy on every one.
- **Not** a Go unit test asserting on generated output. It restates what the generator emitted and goes stale on the next regen.
- The harness expresses happy paths only — `plan → apply → plan-no-drift → destroy`. **A config that must fail at plan cannot be expressed**, and extending the runner to express it is a separate PR. If the repro is a rejected plan, say so and cover what the harness can hold.

Keep the PR to the reproduction. Neighbouring test cases, validators and suite improvements are follow-ups you mention, not commits you add.

### 5. Document the shape in the pattern catalog

Add one entry to [`patterns.md`](../../docs/patterns.md), in the same PR, under Terraform-side or SigNoz-side. Match the existing format: a `### <shape name>` noun phrase, a one-line description of the shape, then `- **Pitfall:**` / `- **Fix:**` / `- **Example:**` citing this issue and the upstream one.

The entry describes the **class**, not the incident. Someone hitting the next instance should recognise it there.

### 6. Write up the suggested upstream fix — don't implement it

Name the file and function, the reason the current code misses the case, and the blast radius: which resources gain or lose what. Give the expected regen delta, labelled as a prediction.

Two things that have both bitten:

- **Look for the sibling bug the fix would expose.** Fixing skaff's map traversal alone would have propagated an existing duplicate-validator bug into every dashboard, so both fixes had to ship together.
- **Predictions run low.** The issue #154 fix restored five validator sites, not the three predicted, and four append sites needed the dedup, not three.

Stop here. Implementing the upstream fix, opening the upstream PR, and turning the repro PR into the fix are all separate asks.

### 7. If asked to carry it upstream

In this order:

1. **Upstream PR first**, and it must be a PR — not a local patch you verify against.
2. Verify by regenerating the provider against the patched tool and confirming the repro flips green.
3. Once upstream releases, rebase the provider PR onto `main`, regenerate in the **same** PR, and take it out of draft.
4. **Pin nothing.** Primus owns the skaff version and resolves the latest release, so a version pin in the provider is a defect. Its download is cached on `[ -f $(SKAFF) ]`, so check what your local binary actually is rather than assuming `latest` re-resolved.

### 8. Hand the reporter-facing text over — don't post it

Write a suggested issue title and comment and give them to the user in chat. **Do not comment on the public issue and do not close it** unless asked, and leave the closing keyword out of the PR body so the merge doesn't auto-close it.

Write it for the reporter, not for a reviewer: thank them, say plainly what the cause turned out to be, correct the filed diagnosis gently if it was wrong, name the fixing PR, show the error they will get instead, and invite a follow-up if their case differs.

## Standing rules

- **Don't fix it in the provider because that's faster.** The next regen wipes it and the next reporter hits it again.
- **A failure you can't explain from the diff is probably upstream drift.** Nothing is pinned — not `foundryctl`, not the SigNoz image, not skaff — deliberately, so server-side contract tightening surfaces early. Fix the rotted fixture in the PR it surfaced on, and never propose pinning.
- **The integration suite is label-gated** on `safe-to-integrate`, so a new scenario has not run against a live SigNoz until someone adds the label. Offline `terraform validate` is the stand-in, and it needs a positive control — deleting a required field must produce an error, or an inert validate passes silently.
- **`terraform validate` can't catch server-side normalization.** An enum carrying several spellings of one value validates, plans and applies, then fails the post-apply read with `Provider produced inconsistent result after apply`. Only the integration suite catches that class.
- PRs follow the [`pull-requests`](../../rules/pull-requests.md) rule. Keep the requester-attribution line first, mirror the template's `####` headings, delete the ones with nothing under them.

## Not this skill

- The upstream spec changed and everything needs regenerating → **spec-syncer**.
- Adding a new `signoz_<name>` resource → **resource-creator**.
- Running the suite, or writing a scenario for its own sake → **integration-tester**.
- A docs or examples gap you have already confirmed → **docs-writer**.
