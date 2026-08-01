from pathlib import Path

import pytest

from fixtures.cli import TERRAFORM, TOOLS

pytest_plugins = [
    "fixtures.signoz",
    "fixtures.channels",
    "fixtures.cli",
    "fixtures.terraform",
]

# tmp/bin/ at the repo root — already gitignored, and shared with the other
# tooling binaries the repo downloads.
DOWNLOAD_PATH = Path(__file__).resolve().parent.parent / "tmp" / "bin"


def pytest_addoption(parser: pytest.Parser):
    parser.addoption(
        "--reuse",
        action="store_true",
        default=False,
        help="Reuse a cached SigNoz environment across runs instead of creating a new one. Run e.g. `pytest --reuse integration/bootstrap/setup.py::test_setup` to stand one up.",
    )
    parser.addoption(
        "--teardown",
        action="store_true",
        default=False,
        help="Tear down the cached SigNoz environment. Run `pytest --teardown integration/bootstrap/setup.py::test_teardown`.",
    )
    parser.addoption(
        "--foundry-binary-path",
        action="store",
        default="foundryctl",
        help="Path to the foundryctl binary used to cast the SigNoz environment.",
    )
    parser.addoption(
        "--tool",
        action="store",
        default=TERRAFORM,
        choices=TOOLS,
        help="Terraform-compatible CLI driven through the CRUD cycle.",
    )
    # Not --version: pytest already owns that flag.
    parser.addoption(
        "--tool-version",
        action="store",
        default="latest",
        help="Version of --tool to download, e.g. 1.9.8. 'latest' resolves the newest release.",
    )
    parser.addoption(
        "--download-path",
        action="store",
        default=str(DOWNLOAD_PATH),
        help="Directory the --tool binary is downloaded into; each tool/version gets its own subdirectory.",
    )
    parser.addoption(
        "--go-binary-path",
        action="store",
        default="go",
        help="Path to the go binary used to build the provider under test.",
    )
    parser.addoption(
        "--license-key",
        action="store",
        default="",
        help="SigNoz license key applied after startup; empty skips license application.",
    )
