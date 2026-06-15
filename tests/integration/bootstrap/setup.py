from http import HTTPStatus

import requests

from fixtures.logger import setup_logger
from fixtures.signoz import SigNoz

logger = setup_logger(__name__)


def test_setup(signoz: SigNoz) -> None:
    """Create (or reuse) the SigNoz environment and confirm it is reachable."""
    response = requests.get(f"{signoz.endpoint}/api/v1/version", timeout=5)
    logger.info("version response: %s", response.status_code)
    assert response.status_code == HTTPStatus.OK


def test_teardown(signoz: SigNoz) -> None:  # noqa: ARG001 — requesting the fixture drives teardown
    """Tear down the cached SigNoz environment (use with --teardown)."""
