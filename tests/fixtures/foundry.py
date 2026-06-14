"""Spin up a SigNoz environment with foundryctl for the test session.

foundryctl (https://github.com/signoz/foundry) is the canonical SigNoz
installer. `foundryctl cast` runs the full pipeline (validate + generate +
deploy); for the docker/compose flavour that brings SigNoz up on :8080. Teardown
is `docker compose down` on the generated compose file under pours/.

Set SIGNOZ_ENDPOINT to point the suite at an already-running instance and skip
foundry entirely (useful for local iteration).
"""

import os
import shutil
import subprocess
import time
from pathlib import Path

import pytest
import requests

from fixtures.logger import setup_logger

logger = setup_logger(__name__)

ENDPOINT = "http://localhost:8080"
FOUNDRYCTL = os.environ.get("FOUNDRYCTL_BIN", "foundryctl")

TESTS_DIR = Path(__file__).resolve().parent.parent
CASTING = TESTS_DIR / "casting.yaml"
POURS = TESTS_DIR / "pours"


def _wait_for_port(endpoint: str, timeout: float = 240.0) -> None:
    """Wait until the SigNoz HTTP port answers at all (any status)."""
    deadline = time.time() + timeout
    last = None

    while time.time() < deadline:
        try:
            requests.get(endpoint, timeout=5)
            return
        except requests.RequestException as err:
            last = err

        time.sleep(3)

    raise TimeoutError(f"{endpoint} did not respond within {timeout}s (last={last})")


def _compose_file() -> Path:
    candidates = sorted(POURS.rglob("compose.yaml"))
    if not candidates:
        raise FileNotFoundError(f"no compose.yaml generated under {POURS}")

    return candidates[0]


def _teardown() -> None:
    try:
        compose = _compose_file()
    except FileNotFoundError:
        return

    subprocess.run(["docker", "compose", "-f", str(compose), "down", "-v"], check=False)
    shutil.rmtree(POURS, ignore_errors=True)


@pytest.fixture(scope="session")
def signoz_endpoint(pytestconfig: pytest.Config):
    """Yield the base URL of a running SigNoz instance for the session."""
    external = os.environ.get("SIGNOZ_ENDPOINT")
    if external:
        logger.info("using existing SigNoz at %s (SIGNOZ_ENDPOINT set)", external)
        _wait_for_port(external)
        yield external
        return

    if shutil.which(FOUNDRYCTL) is None:
        pytest.skip(f"{FOUNDRYCTL} not on PATH; set SIGNOZ_ENDPOINT to use an existing instance")

    logger.info("casting SigNoz with foundryctl (%s)", CASTING)
    subprocess.run([FOUNDRYCTL, "cast", "-f", str(CASTING), "-p", str(POURS)], cwd=TESTS_DIR, check=True)

    try:
        _wait_for_port(ENDPOINT)
        yield ENDPOINT
    finally:
        if pytestconfig.getoption("--keep-env"):
            logger.info("--keep-env set; leaving SigNoz running (teardown: docker compose -f %s down -v)", _compose_file())
        else:
            _teardown()
