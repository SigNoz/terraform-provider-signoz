"""Build the provider under test and drive the Terraform CLI against it.

Terraform is pointed at the freshly built provider binary with a CLI config that
declares a `dev_overrides` block. dev_overrides bypass the registry and the
`terraform init` step entirely — commands resolve the provider straight from the
local build.
"""

import os
import shutil
import subprocess
from pathlib import Path

import pytest

from fixtures.logger import setup_logger
from fixtures.signoz import SigNoz

logger = setup_logger(__name__)

REPO_ROOT = Path(__file__).resolve().parents[2]
EXAMPLES = REPO_ROOT / "examples"

# Provider source address; matches main.go's registry address and the
# `source` used in the generated versions.tf.
PROVIDER_SOURCE = "signoz/signoz"

# Written into each workspace so Terraform resolves signoz_* resources to the
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
def terraform_bin(request: pytest.FixtureRequest) -> str:
    return request.config.getoption("--terraform-binary-path")


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
def workspace(tmp_path: Path, request: pytest.FixtureRequest) -> Path:
    """Stage a single example .tf file into an isolated workspace with provider config.

    Driven by indirect parametrization: `request.param` is the .tf file to stage.
    """
    tf_file: Path = request.param
    shutil.copy(tf_file, tmp_path / tf_file.name)

    (tmp_path / "versions.tf").write_text(VERSIONS_TF)
    return tmp_path


class Terraform:
    """Runs the Terraform CLI in a workspace against the dev-override provider."""

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
        logger.info("terraform %s -> %d", " ".join(args), result.returncode)
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
