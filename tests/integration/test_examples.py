"""Run every resource example file through a full Terraform CRUD cycle.

Each `*.tf` file under `examples/resources/signoz_*` is exercised on its own (so
the six signoz_alert files run one by one): apply (create), confirm there is no
drift (`plan -detailed-exitcode` == 0), then destroy. Every file gets its own
workspace and runs against the real SigNoz instance from the `signoz` fixture.

Data-source examples are not exercised here — they read an object by id, which
does not exist on a fresh instance.
"""

import shutil
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, PROVIDER_SOURCE, Terraform

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("signoz_*/*.tf"))
RESOURCE_IDS = [f"{p.parent.name}/{p.name}" for p in RESOURCE_FILES]

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
    """Stage a single example .tf file into an isolated workspace with provider config."""
    tf_file: Path = request.param
    shutil.copy(tf_file, tmp_path / tf_file.name)

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    return tmp_path


@pytest.mark.parametrize("workspace", RESOURCE_FILES, ids=RESOURCE_IDS, indirect=True)
def test_resource_file_crud(workspace: Path, tf_cli_config: Path, signoz: SigNoz, terraform_bin: str):
    terraform = Terraform(workspace, tf_cli_config, signoz, terraform_bin)

    # Create.
    terraform.apply()

    try:
        # Read: applying again must be a no-op — no drift.
        assert terraform.plan_exit_code() == 0, "drift detected after apply"
    finally:
        # Delete.
        terraform.destroy()
