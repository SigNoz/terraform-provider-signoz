"""Regression test: creating an alert must not churn its rule ID.

The pre-fix provider left `preferred_channels` unknown after Create and ignored `disabled` in the
create payload. Terraform therefore rejected the create result and tainted the freshly created
alert, so the next apply destroyed and recreated it — assigning a brand new rule ID every time.
That is what drove the "rule IDs keep changing" problem for downstream IaC (broken doc links,
lost alert history). See resource/alert.go Create().

This exercises the fix end to end: create succeeds, and a benign update is applied in place with
the rule ID preserved rather than destroy+recreate.
"""

import re
from pathlib import Path

import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES, VERSIONS_TF, Terraform

# A disabled traces alert that leaves the Optional+Computed attributes with no default
# (preferred_channels AND rule_type) unset — the exact shape that tripped the bug: both plan as
# "known after apply" and Create must repopulate them. `disabled = true` also exercises the
# create-payload Disabled fix (and matches how our alerts are shipped: created disabled, enabled later).
ALERT_TF = """\
resource "signoz_alert" "churn" {
  alert          = "rule-id stability regression"
  alert_type     = "TRACES_BASED_ALERT"
  description    = "__DESC__"
  summary        = "rule-id stability regression"
  severity       = "warning"
  version        = "v5"
  schema_version = "v2alpha1"
  disabled       = true

  condition = jsonencode({
    compositeQuery = {
      queryType = "builder"
      panelType = "graph"
      unit      = "ns"
      queries = [{
        type = "builder_query"
        spec = {
          name         = "A"
          signal       = "traces"
          stepInterval = 60
          aggregations = [{ expression = "p99(duration_nano)" }]
          filter       = { expression = "service.name = 'demo'" }
        }
      }]
    }
    selectedQueryName = "A"
    thresholds = {
      kind = "basic"
      spec = [{
        name       = "warning"
        op         = "above"
        matchType  = "at_least_once"
        target     = 10
        targetUnit = "s"
        channels   = ["slack"]
      }]
    }
  })

  evaluation = jsonencode({
    kind = "rolling"
    spec = { evalWindow = "5m", frequency = "1m" }
  })

  notification_settings = {
    group_by = ["service.name"]
    renotify = {
      enabled      = true
      interval     = "30m"
      alert_states = ["firing"]
    }
  }

  labels = {
    team = "regression"
  }
}
"""


def test_alert_rule_id_stable_across_update(
    tmp_path: Path,
    tf_cli_config: Path,
    signoz: SigNoz,
    terraform_bin: str,
    webhook_channels: tuple[str, ...],
):
    workdir = tmp_path / "ws"
    workdir.mkdir()
    (workdir / "versions.tf").write_text(VERSIONS_TF)
    (workdir / "alert.tf").write_text(ALERT_TF.replace("__DESC__", "before"))

    terraform = Terraform(workdir, tf_cli_config, signoz, terraform_bin)

    # Create. Pre-fix this fails: preferred_channels is unknown after apply, so Terraform taints the
    # resource ("invalid result object after apply").
    terraform.apply()
    rule_id = terraform.state_resource_id("signoz_alert")
    assert rule_id, "alert was not created"

    try:
        # A benign change. Pre-fix, the tainted resource is destroyed and recreated with a new id.
        (workdir / "alert.tf").write_text(ALERT_TF.replace("__DESC__", "after"))
        terraform.apply()

        assert terraform.state_resource_id("signoz_alert") == rule_id, (
            "rule ID churned across an update (destroy+recreate instead of in-place)"
        )
    finally:
        terraform.destroy()


def test_alert_enable_disable_toggle(
    tmp_path: Path,
    tf_cli_config: Path,
    signoz: SigNoz,
    terraform_bin: str,
    webhook_channels: tuple[str, ...],
):
    """Enabling a disabled alert (disabled true -> false) must be in place, not a recreate.

    This is the "ship the alert disabled, enable it later" workflow. Pre-fix, Create ignored disabled
    in the payload, so the alert never actually landed disabled and the enable path was unreliable.
    """
    workdir = tmp_path / "ws"
    workdir.mkdir()
    (workdir / "versions.tf").write_text(VERSIONS_TF)
    (workdir / "alert.tf").write_text(ALERT_TF.replace("__DESC__", "toggle"))  # disabled = true

    terraform = Terraform(workdir, tf_cli_config, signoz, terraform_bin)

    terraform.apply()
    created = terraform.state_resource("signoz_alert")
    assert created["disabled"] is True, "alert was not created disabled"
    rule_id = created["id"]

    try:
        (workdir / "alert.tf").write_text(
            ALERT_TF.replace("__DESC__", "toggle").replace("disabled       = true", "disabled       = false")
        )
        terraform.apply()

        after = terraform.state_resource("signoz_alert")
        assert after["disabled"] is False, "enabling the alert did not take effect"
        assert after["id"] == rule_id, "enabling the alert churned the rule ID"
    finally:
        terraform.destroy()


def test_alert_import_roundtrip(
    tmp_path: Path,
    tf_cli_config: Path,
    signoz: SigNoz,
    terraform_bin: str,
    webhook_channels: tuple[str, ...],
):
    """After importing an existing alert by id, reconciling must be in place — not a destroy+recreate."""
    workdir = tmp_path / "ws"
    workdir.mkdir()
    (workdir / "versions.tf").write_text(VERSIONS_TF)
    (workdir / "alert.tf").write_text(ALERT_TF.replace("__DESC__", "import"))

    terraform = Terraform(workdir, tf_cli_config, signoz, terraform_bin)

    terraform.apply()
    rule_id = terraform.state_resource_id("signoz_alert")

    try:
        terraform.state_rm("signoz_alert.churn")
        terraform.import_resource("signoz_alert.churn", rule_id)
        # Applying against the imported state must reconcile in place, preserving the id.
        terraform.apply()
        assert terraform.state_resource_id("signoz_alert") == rule_id, (
            "import round-trip replaced the resource (new rule ID)"
        )
    finally:
        terraform.destroy()


# The churn (unknown-value-after-apply) affected every alert type, not just traces. Assert each shipped
# example creates without the "invalid result object after apply" error. Channel names are rewritten to a
# channel that exists in the test env (examples reference environment-specific channels).
_ALERT_EXAMPLES = sorted((EXAMPLES / "resources" / "signoz_alert").glob("*.tf"))
_CREATE_SKIP = {"resource_anomaly.tf": 'API rejects rule_type "anomaly_rule" (unrelated to this fix)'}


@pytest.mark.parametrize(
    "example",
    [
        pytest.param(
            p,
            id=p.name,
            marks=([pytest.mark.skip(reason=_CREATE_SKIP[p.name])] if p.name in _CREATE_SKIP else []),
        )
        for p in _ALERT_EXAMPLES
    ],
)
def test_alert_example_creates_without_unknown_values(
    example: Path,
    tmp_path: Path,
    tf_cli_config: Path,
    signoz: SigNoz,
    terraform_bin: str,
    webhook_channels: tuple[str, ...],
):
    src = re.sub(r"channels(\s*)=(\s*)\[[^\]]*\]", r'channels\1=\2["slack"]', example.read_text())
    workdir = tmp_path / "ws"
    workdir.mkdir()
    (workdir / "versions.tf").write_text(VERSIONS_TF)
    (workdir / example.name).write_text(src)

    terraform = Terraform(workdir, tf_cli_config, signoz, terraform_bin)

    # Pre-fix this errors with "invalid result object after apply" (unknown preferred_channels/rule_type).
    terraform.apply()
    try:
        assert terraform.state_resource_id("signoz_alert")
    finally:
        terraform.destroy()
