import json
from pathlib import Path

import jsonpatch
import pytest

from fixtures.signoz import SigNoz
from fixtures.tool import TESTDATA, VERSIONS_TF, Tool

# resources/signoz_<name>/<NN>/ — each two-digit dir is one scenario.
SCENARIOS = sorted(p for p in (TESTDATA / "resources").glob("signoz_*/[0-9][0-9]") if p.is_dir())


@pytest.mark.parametrize("scenario", SCENARIOS, ids=[f"{s.parent.name}/{s.name}" for s in SCENARIOS])
def test_scenario_lifecycle(scenario: Path, tmp_path: Path, tool_config: Path, signoz: SigNoz, tool_bin: str, webhook_channels: tuple[str, ...]):
    bases = [p for p in scenario.iterdir() if p.name.endswith((".tf", ".tf.json")) and not p.name.endswith("-jsonpatch.json")]
    assert len(bases) == 1, f"{scenario}: expected exactly one base .tf/.tf.json, found {sorted(p.name for p in bases)}"

    base = bases[0]
    patches = sorted(scenario.glob("*-jsonpatch.json"))
    is_json = base.name.endswith(".tf.json")

    assert is_json or not patches, f"{scenario}: JSON patches require a .tf.json base, got {base.name}"

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    tool = Tool(tmp_path, tool_config, signoz, tool_bin)

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
        assert tool.plan_exit_code() == 2, "expected a create on the first plan"
        tool.apply()
        assert tool.plan_exit_code() == 0, "drift after initial apply"

        for patch in patches:
            body = jsonpatch.apply_patch(body, json.loads(patch.read_text()))
            named[rname] = body
            config.write_text(json.dumps(doc, indent=2))

            assert tool.plan_exit_code() == 2, f"{patch.name}: expected the edit to change the plan"
            tool.apply()
            assert tool.plan_exit_code() == 0, f"{patch.name}: drift after applying the edit"
    finally:
        tool.destroy()
