"""Build the provider under test and drive a Terraform-compatible CLI against it.

The CLI — Terraform or OpenTofu, supplied by the `tool_bin` fixture — is pointed
at the freshly built provider binary with a CLI config that declares a
`dev_overrides` block. dev_overrides bypass the registry and the `init` step
entirely: commands resolve the provider straight from the local build. Both CLIs
read the config from `TF_CLI_CONFIG_FILE` and normalize the `signoz/signoz`
source address against their own default registry host, so one config serves
both.
"""

import os
import shutil
import subprocess
from collections.abc import Callable
from pathlib import Path

import pytest

from fixtures.logger import setup_logger
from fixtures.signoz import SigNoz

logger = setup_logger(__name__)

REPO_ROOT = Path(__file__).resolve().parents[2]
EXAMPLES = REPO_ROOT / "examples"
# Edge-case configs that exercise the provider beyond the user-facing examples;
# same layout as examples/ (resources/signoz_<name>/*.tf).
TESTDATA = REPO_ROOT / "tests" / "testdata"

# Provider source address, left unqualified on purpose: each CLI normalizes it
# against its own default registry host, so the same string matches the
# dev_overrides key under both Terraform and OpenTofu.
PROVIDER_SOURCE = "signoz/signoz"

# Written into each workspace so the CLI resolves signoz_* resources to the
# dev-overridden provider; the provider reads endpoint/token from the env.
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


@pytest.fixture(scope="session")
def provider_dir(request: pytest.FixtureRequest, tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Build the provider binary into a directory for Terraform dev_overrides."""
    go = request.config.getoption("--go-binary-path")
    out = tmp_path_factory.mktemp("provider-bin")
    binary = out / "terraform-provider-signoz"

    logger.info("building provider with %s -> %s", go, binary)
    subprocess.run([go, "build", "-o", str(binary), "."], cwd=REPO_ROOT, check=True)

    return out


@pytest.fixture(scope="session")
def tf_cli_config(provider_dir: Path, tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Write a Terraform CLI config that dev-overrides the provider to the local build."""
    cfg = tmp_path_factory.mktemp("tf-cli") / "dev.tfrc"
    cfg.write_text(f'provider_installation {{\n  dev_overrides {{\n    "{PROVIDER_SOURCE}" = "{provider_dir}"\n  }}\n  direct {{}}\n}}\n')

    return cfg


@pytest.fixture
def workspace(tmp_path_factory: pytest.TempPathFactory) -> Callable[[Path], Path]:
    """Return a factory that stages an example .tf file into its own workspace with provider config."""

    def stage(tf_file: Path) -> Path:
        workdir = tmp_path_factory.mktemp(f"{tf_file.parent.name}-{tf_file.stem}")

        shutil.copy(tf_file, workdir / tf_file.name)
        (workdir / "versions.tf").write_text(VERSIONS_TF)
        return workdir

    return stage


class Terraform:
    """Runs a Terraform-compatible CLI in a workspace against the dev-override provider."""

    def __init__(self, workdir: Path, cli_config: Path, signoz: SigNoz, binary: str = "terraform"):
        self.workdir = workdir
        self.binary = binary
        self.env = {
            **os.environ,
            "TF_CLI_CONFIG_FILE": str(cli_config),
            "TF_IN_AUTOMATION": "1",
            "SIGNOZ_ENDPOINT": signoz.endpoint,
            "SIGNOZ_ACCESS_TOKEN": signoz.access_token,
        }

    def _run(self, *args: str) -> subprocess.CompletedProcess:
        # dev_overrides make `init` unnecessary (and it would error on the
        # missing dependency lock), so commands run directly.
        result = subprocess.run(
            [self.binary, *args, "-no-color"],
            cwd=self.workdir,
            env=self.env,
            text=True,
            capture_output=True,
        )
        logger.info("%s %s -> %d", Path(self.binary).name, " ".join(args), result.returncode)
        return result

    def apply(self) -> subprocess.CompletedProcess:
        result = self._run("apply", "-auto-approve")
        assert result.returncode == 0, f"apply failed:\n{result.stdout}\n{result.stderr}"
        return result

    def plan_exit_code(self) -> int:
        # -detailed-exitcode: 0 = no changes, 1 = error, 2 = changes (drift).
        result = self._run("plan", "-detailed-exitcode")
        assert result.returncode in (0, 2), f"plan errored:\n{result.stdout}\n{result.stderr}"
        return result.returncode

    def destroy(self) -> subprocess.CompletedProcess:
        result = self._run("destroy", "-auto-approve")
        assert result.returncode == 0, f"destroy failed:\n{result.stdout}\n{result.stderr}"
        return result
