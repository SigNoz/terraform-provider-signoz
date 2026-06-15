from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, Terraform

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("signoz_*/*.tf"))
RESOURCE_IDS = [f"{p.parent.name}/{p.name}" for p in RESOURCE_FILES]


@pytest.mark.parametrize("tf_file", RESOURCE_FILES, ids=RESOURCE_IDS)
def test_resource_file_crud(tf_file: Path, workspace: Callable[[Path], Path], tf_cli_config: Path, signoz: SigNoz, terraform_bin: str, webhook_channels: tuple[str, ...]):
    terraform = Terraform(workspace(tf_file), tf_cli_config, signoz, terraform_bin)

    # Create.
    terraform.apply()

    try:
        # Read: applying again must be a no-op — no drift.
        assert terraform.plan_exit_code() == 0, "drift detected after apply"
    finally:
        # Delete.
        terraform.destroy()
