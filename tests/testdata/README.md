# Test data

Terraform configs that exercise the provider beyond the curated, user-facing
[`examples/`](../../examples). Same top-level layout — `resources/signoz_<name>/`
— so the integration suite can drive every config here. There are two kinds of
fixture under each `signoz_<name>/` directory:

## Single-shot configs — `signoz_<name>/*.tf`

Flat `.tf` files, one resource each, run by
[`test_testdata.py`](../integration/tests/test_testdata.py) through a single
Create → Read (no-drift) → Delete cycle. These are edge cases, not
documentation: unusual-but-valid field combinations, every query kind, every
evaluation window, multiple thresholds, and so on.

## Edit scenarios — `signoz_<name>/NN/`

Two-digit directories are ordered edit scenarios, run by
[`test_editing.py`](../integration/tests/test_editing.py):

```
signoz_rule/01/
  01-resource.tf.json   # base config
  02-jsonpatch.json     # RFC 6902 patch applied on top of the base
  03-jsonpatch.json     # RFC 6902 patch applied on top of the previous state
```

The runner creates the base (plan shows a create, apply, re-plan is clean), then
applies each `NN-jsonpatch.json` in ascending order — re-plan must show changes,
apply, re-plan must be clean — and finally destroys. This is the end-to-end
editing path: create, no-drift, edit, drift, converge, repeat, delete.

The base is `01-resource.tf(.json)`. A JSON Patch (RFC 6902) needs a JSON target,
so any scenario with patches uses a **`.tf.json`** base (Terraform JSON syntax,
read natively by Terraform); patch `path`s are relative to the single resource's
body (e.g. `/condition/thresholds/basic/spec/0/target`). A patch-free scenario
may use a plain HCL `.tf` base.

## Channels

Channels referenced by `thresholds[*].channels` must be seeded by the suite —
`slack` and `pagerduty` are (see `fixtures/channels.py`).
