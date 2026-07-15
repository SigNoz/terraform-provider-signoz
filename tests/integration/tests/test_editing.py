"""End-to-end edit scenarios driven by ordered step files.

Each numbered directory under testdata/resources/signoz_<name>/ is one scenario:

    signoz_rule/01/
      01-resource.tf.json   # base config (Terraform JSON syntax)
      02-jsonpatch.json     # RFC 6902 patch applied on top of the base
      03-jsonpatch.json     # RFC 6902 patch applied on top of the previous state

The runner creates the base (plan shows a create, apply, re-plan is clean), then
applies each jsonpatch in order — re-plan must show changes, apply, re-plan must
be clean — and finally destroys. A scenario with no patches degenerates to the
plain create -> no-drift -> destroy cycle, and its base may be plain HCL (.tf);
scenarios with patches need a .tf.json base for the patch to target.
"""

import json
from pathlib import Path

import jsonpatch
import pytest

from fixtures.signoz import SigNoz
from fixtures.terraform import TESTDATA, VERSIONS_TF, Terraform

# resources/signoz_<name>/<NN>/ — two-digit dirs are edit scenarios.
SCENARIOS = sorted(p for p in (TESTDATA / "resources").glob("signoz_*/[0-9][0-9]") if p.is_dir())


def _id(scenario: Path) -> str:
    return f"{scenario.parent.name}/{scenario.name}"


def _base_step(scenario: Path) -> Path:
    bases = sorted(scenario.glob("*-resource.tf*"))
    assert len(bases) == 1, f"{scenario}: expected exactly one *-resource.tf(.json), found {[p.name for p in bases]}"

    return bases[0]


def _patch_steps(scenario: Path) -> list[Path]:
    return sorted(scenario.glob("*-jsonpatch.json"))


@pytest.mark.parametrize("scenario", SCENARIOS, ids=_id)
def test_edit_scenario(scenario: Path, tmp_path: Path, tf_cli_config: Path, signoz: SigNoz, terraform_bin: str, webhook_channels: tuple[str, ...]):
    base = _base_step(scenario)
    patches = _patch_steps(scenario)
    is_json = base.name.endswith(".tf.json")

    assert is_json or not patches, f"{scenario}: JSON patches require a .tf.json base, got {base.name}"

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    terraform = Terraform(tmp_path, tf_cli_config, signoz, terraform_bin)

    if is_json:
        doc = json.loads(base.read_text())
        (rtype, named), = doc["resource"].items()
        (rname, body), = named.items()

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
