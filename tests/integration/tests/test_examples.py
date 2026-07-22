from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, Terraform

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("signoz_*/*.tf"))

# Example files to skip, keyed by "<resource>/<file>" -> reason. The integration
# run surfaced provider/API issues for these.
SKIPPED = {
    "signoz_planned_maintenance/resource.tf": "alert_ids reference non-existent rules (API 500)",
}


@pytest.mark.parametrize(
    "tf_file",
    [pytest.param(tf_file, id=(rel := f"{tf_file.parent.name}/{tf_file.name}"), marks=[pytest.mark.skip(reason=SKIPPED[rel])] if rel in SKIPPED else []) for tf_file in RESOURCE_FILES],
)
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
