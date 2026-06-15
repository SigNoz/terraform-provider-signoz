# Integration tests

End-to-end tests that spin up a real SigNoz instance, run every example in [`examples/`](../examples) through a full Terraform CRUD cycle, and assert there is no drift.

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
