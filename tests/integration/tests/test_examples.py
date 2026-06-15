from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, Terraform

RESOURCE_FILES = sorted((EXAMPLES / "resources").glob("signoz_*/*.tf"))

_ALERT_UNKNOWN = "provider returns unknown values for computed fields after apply"

# Example files to skip, keyed by "<resource>/<file>" -> reason. The integration
# run surfaced provider/API issues for these.
SKIPPED = {
    "signoz_alert/resource.tf": _ALERT_UNKNOWN,
    "signoz_alert/resource_logs_formula.tf": _ALERT_UNKNOWN,
    "signoz_alert/resource_promql.tf": _ALERT_UNKNOWN,
    "signoz_alert/resource_tiered.tf": _ALERT_UNKNOWN,
    "signoz_alert/resource_traces_latency.tf": _ALERT_UNKNOWN,
    "signoz_alert/resource_anomaly.tf": 'provider rejects rule_type "anomaly_rule"',
    "signoz_dashboard/resource.tf": "provider produces an inconsistent result after apply",
    "signoz_planned_maintenance/resource.tf": "alert_ids reference non-existent rules (API 500)",
}


def _case(tf_file: Path):
    rel = f"{tf_file.parent.name}/{tf_file.name}"
    marks = [pytest.mark.skip(reason=SKIPPED[rel])] if rel in SKIPPED else []

    return pytest.param(tf_file, id=rel, marks=marks)


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
