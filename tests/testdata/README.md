# Test data

Terraform configs that exercise the provider beyond the curated, user-facing
[`examples/`](../../examples). Same layout — `resources/signoz_<name>/*.tf` — so
[`test_testdata.py`](../integration/tests/test_testdata.py) can run every file
here through the same Create → Read (no-drift) → Delete cycle as
`test_examples.py`.

These are edge cases, not documentation: unusual-but-valid field combinations,
every query kind, every evaluation window, multiple thresholds, and so on. Keep
each file to a single resource so a failure points at one config.

Channels referenced by `thresholds[*].channels` must be seeded by the suite —
`slack` and `pagerduty` are (see `fixtures/channels.py`).
