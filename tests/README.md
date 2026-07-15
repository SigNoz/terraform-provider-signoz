# Integration tests

End-to-end tests that spin up a real SigNoz instance and drive Terraform against a locally built provider, asserting there is no drift. Two suites run:

- [`test_examples.py`](integration/tests/test_examples.py) — every example in [`examples/`](../examples) through a full Terraform CRUD cycle.
- [`test_testdata.py`](integration/tests/test_testdata.py) — every scenario in [`testdata/`](testdata) through a full lifecycle, including edits (see [Test data](#test-data)).

## Requirements

- [`foundryctl`](https://github.com/signoz/foundry) on `PATH` (or set `SIGNOZ_ENDPOINT`)
- Docker (for the Compose deployment)
- Terraform CLI
- Go (to build the provider under test)
- Python ≥ 3.11 with [`uv`](https://docs.astral.sh/uv/)

## Running

```sh
cd tests
uv sync
uv run pytest -vv
```

## Test data

`testdata/` holds Terraform configs that exercise the provider beyond the curated, user-facing [`examples/`](../examples). Same top-level layout — `resources/signoz_<name>/` — so the suite can drive every config here.

Each two-digit directory under `signoz_<name>/` is one scenario: a single base config plus, optionally, ordered JSON-patch edits. All scenarios are run by [`test_testdata.py`](integration/tests/test_testdata.py) through a full lifecycle.

```
signoz_rule/00/
  01-<name>.tf            # base config, no edits

signoz_rule/01/
  01-<name>.tf.json       # base config
  02-jsonpatch.json       # RFC 6902 patch applied on top of the base
  03-jsonpatch.json       # RFC 6902 patch applied on top of the previous state
```

The runner creates the base (plan shows a create, apply, re-plan is clean), applies each `NN-jsonpatch.json` in ascending order — re-plan must show changes, apply, re-plan must be clean — and finally destroys. A scenario with no patches is just create → no-drift → destroy; the full sequence is the end-to-end editing path: create, no-drift, edit, drift, converge, repeat, delete.

Naming:

- The base is the single `01-<name>.tf` or `01-<name>.tf.json` in the directory; the `<name>` says what the scenario exercises.
- A JSON Patch (RFC 6902) needs a JSON target, so any scenario with patches uses a **`.tf.json`** base (Terraform JSON syntax, read natively by Terraform). Patch `path`s are relative to the single resource's body (e.g. `/condition/thresholds/basic/spec/0/target`). A patch-free scenario may use a plain HCL `.tf` base.

Channels referenced by `thresholds[*].channels` must be seeded by the suite — `slack` and `pagerduty` are (see [`fixtures/channels.py`](fixtures/channels.py)).
