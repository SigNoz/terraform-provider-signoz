# terraform-provider-signoz

A [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) provider for [SigNoz](https://signoz.io). It is **generated, not hand-written**: the [`SigNoz/skaff`](https://github.com/SigNoz/skaff) codegen tool reads the SigNoz OpenAPI spec plus a per-resource config (`skaff.yml`) and emits the `zz_generated_*` files. Adding a resource means *declaring* it and *running the pipeline* — you rarely write resource logic by hand.

## Architecture

@.claude/docs/architecture.md

## Primus

skaff and the Go gates run through **primus**. If it's not set up (`PRIMUS_HOME` unset), use the **primus-setter** skill.

## Conventions

@.claude/rules

## Skills

| Task | Skill |
|------|-------|
| Add / scaffold a new `signoz_<name>` resource + data source from an API | **resource-creator** |
| Regenerate everything from the latest upstream spec + open a catch-up PR | **spec-syncer** |
| Regenerate or validate registry docs, add examples, write a guide | **docs-writer** |
| Run the e2e suite, test against a real SigNoz, add a testdata scenario | **integration-tester** |
| Install primus tooling | **primus-setter** |

Reach for the skill before hand-writing — it carries the runbook, the flags, the footguns, and the design/history references.
