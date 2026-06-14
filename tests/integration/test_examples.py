"""Run every resource example through a full Terraform CRUD cycle.

For each `examples/resources/signoz_*` directory: apply (create), confirm there
is no drift (`plan -detailed-exitcode` == 0), then destroy. Each example is its
own workspace and runs against the real SigNoz instance from the `signoz` fixture.

Data-source examples are not exercised here — they read an object by id, which
does not exist on a fresh instance.
"""

import shutil
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, PROVIDER_SOURCE, Terraform

RESOURCE_DIRS = sorted(p for p in (EXAMPLES / "resources").glob("signoz_*") if p.is_dir())
RESOURCE_IDS = [p.name for p in RESOURCE_DIRS]

VERSIONS_TF = f"""\
terraform {{
  required_providers {{
    signoz = {{
      source = "{PROVIDER_SOURCE}"
    }}
  }}
}}

provider "signoz" {{}}
"""


@pytest.fixture
def workspace(tmp_path: Path, request: pytest.FixtureRequest) -> Path:
    """Stage one example's *.tf into an isolated workspace with provider config."""
    example_dir: Path = request.param
    for tf in example_dir.glob("*.tf"):
        shutil.copy(tf, tmp_path / tf.name)

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    return tmp_path


@pytest.mark.parametrize("workspace", RESOURCE_DIRS, ids=RESOURCE_IDS, indirect=True)
def test_resource_example_crud(workspace: Path, tf_cli_config: Path, signoz: SigNoz):
    terraform = Terraform(workspace, tf_cli_config, signoz)

    # Create.
    terraform.apply()

    try:
        # Read: applying again must be a no-op — no drift.
        assert terraform.plan_exit_code() == 0, "drift detected after apply"
    finally:
        # Delete.
        terraform.destroy()
