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
3. **Exercise** — [`fixtures/terraform.py`](fixtures/terraform.py) builds the provider
   and points Terraform at it with a `dev_overrides` CLI config (no registry, no
   `init`). [`integration/tests/test_examples.py`](integration/tests/test_examples.py) then stages
   each `examples/resources/signoz_*` directory into its own workspace and runs
   `apply` → `plan -detailed-exitcode` (must be `0` — no drift) → `destroy`.

   Data-source examples aren't exercised (they read an object by id, which doesn't
   exist on a fresh instance). True update (mutating re-apply) is a per-resource
   follow-up; today's cycle is create → no-drift → destroy.

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

- `--reuse` — reuse a cached SigNoz environment across runs (skip create + teardown).
  Stand one up once with
  `uv run pytest --reuse integration/bootstrap/setup.py::test_setup`, then iterate.
- `--teardown` — tear the cached environment back down:
  `uv run pytest --teardown integration/bootstrap/setup.py::test_teardown`.
- `--foundry-binary-path` / `--terraform-binary-path` / `--go-binary-path` — point at
  specific binaries (default `foundryctl` / `terraform` / `go`).
- `SIGNOZ_ENDPOINT=http://...` — test against an already-running SigNoz and skip foundry.

> The root user credentials in `casting.yaml` and `fixtures/signoz.py`
> (`ROOT_EMAIL` / `ROOT_PASSWORD`) must stay in sync.
