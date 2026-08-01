import hashlib
import io
import platform
import re
import stat
import zipfile
from pathlib import Path

import pytest
import requests

from fixtures.logger import setup_logger

logger = setup_logger(__name__)

TERRAFORM = "terraform"
OPENTOFU = "opentofu"
TOOLS = (TERRAFORM, OPENTOFU)

# The executable packaged inside each tool's release archive.
BINARIES = {TERRAFORM: "terraform", OPENTOFU: "tofu"}

# A release is x.y.z; the OpenTofu index also lists -rc / -beta builds.
STABLE = re.compile(r"^\d+\.\d+\.\d+$")


def platform_suffix() -> str:
    """Return the `<os>_<arch>` release-artifact suffix. Both tools name artifacts the same way."""
    systems = {"darwin": "darwin", "linux": "linux"}
    machines = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}

    system, machine = platform.system().lower(), platform.machine().lower()
    assert system in systems, f"unsupported OS {system!r}; the suite runs on {sorted(systems)}"
    assert machine in machines, f"unsupported architecture {machine!r}; the suite runs on {sorted(set(machines.values()))}"

    return f"{systems[system]}_{machines[machine]}"


def resolve(tool: str, version: str) -> str:
    """Resolve "latest" to a concrete release; any other value is taken as given."""
    if version != "latest":
        return version

    if tool == TERRAFORM:
        response = requests.get("https://checkpoint-api.hashicorp.com/v1/check/terraform", timeout=30)
        assert response.status_code == 200, response.text
        resolved = response.json()["current_version"]
    else:
        # The OpenTofu index is neither sorted nor stable-only, so take the highest x.y.z.
        response = requests.get("https://get.opentofu.org/tofu/api.json", timeout=30)
        assert response.status_code == 200, response.text
        resolved = max((v["id"] for v in response.json()["versions"] if STABLE.match(v["id"])), key=lambda v: tuple(int(part) for part in v.split(".")))

    logger.info("resolved latest %s -> %s", tool, resolved)
    return resolved


def download(tool: str, version: str, into: Path) -> Path:
    """Download `tool` at `version` into `into` and return the executable.

    An already-downloaded binary is reused as-is: the caller keys the directory
    on tool + version, so a hit is always the build that was asked for.
    """
    binary = into / BINARIES[tool]
    if binary.exists():
        logger.info("reusing %s %s at %s", tool, version, binary)
        return binary

    suffix = platform_suffix()
    if tool == TERRAFORM:
        base = f"https://releases.hashicorp.com/terraform/{version}"
        archive, checksums = f"terraform_{version}_{suffix}.zip", f"terraform_{version}_SHA256SUMS"
    else:
        base = f"https://github.com/opentofu/opentofu/releases/download/v{version}"
        archive, checksums = f"tofu_{version}_{suffix}.zip", f"tofu_{version}_SHA256SUMS"

    logger.info("downloading %s/%s", base, archive)
    zipped = requests.get(f"{base}/{archive}", timeout=300)
    assert zipped.status_code == 200, f"{base}/{archive}: HTTP {zipped.status_code}"

    # The binary gets executed, so check it against the published sums before unpacking.
    sums = requests.get(f"{base}/{checksums}", timeout=60)
    assert sums.status_code == 200, f"{base}/{checksums}: HTTP {sums.status_code}"

    expected = next(digest for digest, name in (line.split() for line in sums.text.splitlines() if line.strip()) if name.lstrip("*") == archive)
    actual = hashlib.sha256(zipped.content).hexdigest()
    assert actual == expected, f"{archive}: checksum mismatch (expected {expected}, got {actual})"

    into.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(io.BytesIO(zipped.content)) as unpacked:
        unpacked.extract(BINARIES[tool], into)

    binary.chmod(binary.stat().st_mode | stat.S_IEXEC)
    logger.info("installed %s %s at %s", tool, version, binary)

    return binary


@pytest.fixture(scope="session")
def tool_bin(request: pytest.FixtureRequest) -> str:
    """Path to the Terraform-compatible CLI binary the suite drives."""
    tool = request.config.getoption("--tool")
    version = resolve(tool, request.config.getoption("--tool-version"))
    # Absolute: the CLI is executed with cwd set to a workspace temp dir, so a
    # relative --download-path would resolve against the wrong directory.
    into = Path(request.config.getoption("--download-path")).resolve() / tool / version

    return str(download(tool, version, into))
