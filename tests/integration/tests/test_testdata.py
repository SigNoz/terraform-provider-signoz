"""End-to-end lifecycle tests for every scenario under testdata/.

Each numbered directory under testdata/resources/signoz_<name>/ is one scenario:
a single base config plus, optionally, ordered JSON-patch edits.

    signoz_rule/00/
      01-<name>.tf            # base config (HCL or Terraform JSON)

    signoz_rule/01/
      01-<name>.tf.json       # base config (Terraform JSON syntax)
      02-jsonpatch.json       # RFC 6902 patch applied on top of the base
      03-jsonpatch.json       # RFC 6902 patch applied on top of the previous state

For each scenario the runner creates the base (plan shows a create, apply,
re-plan is clean), applies each jsonpatch in ascending order — re-plan must show
changes, apply, re-plan must be clean — and finally destroys. A scenario with no
patches is just create -> no-drift -> destroy; its base may be plain HCL (.tf).
A JSON patch needs a JSON target, so a scenario with patches needs a .tf.json
base (Terraform reads .tf.json natively).
"""

import json
from pathlib import Path

import jsonpatch
import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import TESTDATA, VERSIONS_TF, Terraform

# resources/signoz_<name>/<NN>/ — each two-digit dir is one scenario.
SCENARIOS = sorted(p for p in (TESTDATA / "resources").glob("signoz_*/[0-9][0-9]") if p.is_dir())


def _id(scenario: Path) -> str:
    return f"{scenario.parent.name}/{scenario.name}"


def _base_step(scenario: Path) -> Path:
    bases = [p for p in scenario.iterdir() if p.name.endswith((".tf", ".tf.json")) and not p.name.endswith("-jsonpatch.json")]
    assert len(bases) == 1, f"{scenario}: expected exactly one base .tf/.tf.json, found {sorted(p.name for p in bases)}"

    return bases[0]


def _patch_steps(scenario: Path) -> list[Path]:
    return sorted(scenario.glob("*-jsonpatch.json"))


@pytest.mark.parametrize("scenario", SCENARIOS, ids=_id)
def test_scenario_lifecycle(scenario: Path, tmp_path: Path, tf_cli_config: Path, signoz: SigNoz, terraform_bin: str, webhook_channels: tuple[str, ...]):
    base = _base_step(scenario)
    patches = _patch_steps(scenario)
    is_json = base.name.endswith(".tf.json")

    assert is_json or not patches, f"{scenario}: JSON patches require a .tf.json base, got {base.name}"

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    terraform = Terraform(tmp_path, tf_cli_config, signoz, terraform_bin)

    if is_json:
        doc = json.loads(base.read_text())
        ((_rtype, named),) = doc["resource"].items()
        ((rname, body),) = named.items()

        config = tmp_path / "resource.tf.json"
        config.write_text(json.dumps(doc, indent=2))
    else:
        config = tmp_path / base.name
        config.write_text(base.read_text())

    try:
        # Create: the first plan is the create, apply, re-plan must be clean.
        assert terraform.plan_exit_code() == 2, "expected a create on the first plan"
        terraform.apply()
        assert terraform.plan_exit_code() == 0, "drift after initial apply"

        # Each patch is one edit: re-plan shows changes, apply, re-plan clean.
        for patch in patches:
            body = jsonpatch.apply_patch(body, json.loads(patch.read_text()))
            named[rname] = body
            config.write_text(json.dumps(doc, indent=2))

            assert terraform.plan_exit_code() == 2, f"{patch.name}: expected the edit to change the plan"
            terraform.apply()
            assert terraform.plan_exit_code() == 0, f"{patch.name}: drift after applying the edit"
    finally:
        terraform.destroy()
