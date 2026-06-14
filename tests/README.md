# Integration tests

End-to-end tests that spin up a real SigNoz instance, run every example in
[`examples/`](../examples) through a full Terraform CRUD cycle, and assert there
is no drift.

Modelled on the pytest harness in the [SigNoz repo](https://github.com/SigNoz/signoz/tree/main/tests).

## How it works

1. **Spin up** — [`fixtures/foundry.py`](fixtures/foundry.py) runs `foundryctl cast`
   against [`casting.yaml`](casting.yaml), which deploys SigNoz with Docker Compose
   and provisions a **root user** on first boot (`SIGNOZ_USER_ROOT_*`). Teardown is
   `docker compose down` on the generated `pours/` compose file.
2. **Authenticate** — [`fixtures/signoz.py`](fixtures/signoz.py) logs in as the root
   user, creates a **service account** with the `signoz-admin` role, and mints an
   **API key**. That key is the SigNoz access token the provider authenticates with.
3. **Exercise** *(planned, next step)* — build the provider, point Terraform at it
   with a `dev_overrides` CLI config, and for each example run
   `apply` → `plan -detailed-exitcode` (must be `0` — no drift) → `destroy`.

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

Useful flags / env:

- `SIGNOZ_ENDPOINT=http://...` — test against an already-running SigNoz and skip foundry.
- `--keep-env` — leave the environment running after the run (skip teardown).
- `FOUNDRYCTL_BIN=/path/to/foundryctl` — override the foundry binary.

> The root user credentials in `casting.yaml` and `fixtures/signoz.py`
> (`ROOT_EMAIL` / `ROOT_PASSWORD`) must stay in sync.
