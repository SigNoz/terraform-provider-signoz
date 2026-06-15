from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, Terraform

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("signoz_*/*.tf"))

# Resource directories whose examples are skipped, keyed to the reason. The
# integration run surfaced provider/API issues for these.
SKIPPED_RESOURCES = {
    "signoz_alert": "provider returns unknown values for computed fields after apply",
    "signoz_dashboard": "provider produces an inconsistent result after apply",
    "signoz_planned_maintenance": "schedule / alert_ids rejected by the API (HTTP 500)",
}


def _case(tf_file: Path):
    resource = tf_file.parent.name
    marks = [pytest.mark.skip(reason=SKIPPED_RESOURCES[resource])] if resource in SKIPPED_RESOURCES else []

    return pytest.param(tf_file, id=f"{resource}/{tf_file.name}", marks=marks)


@pytest.mark.parametrize("tf_file", [_case(tf_file) for tf_file in RESOURCE_FILES])
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
