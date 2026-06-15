"""Seed the notification channels the example alerts and route policies reference.

The examples route to channels named "slack" and "pagerduty". SigNoz validates
that referenced channels exist, so they must be created before the CRUD cycle.
"""

import pytest
import requests

from fixtures.logger import setup_logger
from fixtures.signoz import SigNoz

logger = setup_logger(__name__)

WEBHOOK_CHANNELS = ("slack", "pagerduty")
WEBHOOK_URL = "https://example.com/webhook"

_TIMEOUT = 10


def ensure_webhook_channel(signoz: SigNoz, name: str) -> None:
    """Create a webhook notification channel by name if it does not already exist."""
    headers = {"SIGNOZ-API-KEY": signoz.access_token}

    listing = requests.get(f"{signoz.endpoint}/api/v1/channels", headers=headers, timeout=_TIMEOUT)
    listing.raise_for_status()
    existing = {channel.get("name") for channel in (listing.json().get("data") or [])}
    if name in existing:
        logger.info("notification channel %s already exists", name)
        return

    resp = requests.post(
        f"{signoz.endpoint}/api/v1/channels",
        json={"name": name, "webhook_configs": [{"send_resolved": True, "url": WEBHOOK_URL, "http_config": {}}]},
        headers=headers,
        timeout=_TIMEOUT,
    )
    assert resp.status_code == 201, resp.text
    logger.info("created webhook notification channel %s", name)


@pytest.fixture(scope="session")
def webhook_channels(signoz: SigNoz) -> tuple[str, ...]:
    """Ensure the 'slack' and 'pagerduty' webhook channels exist for the examples."""
    for name in WEBHOOK_CHANNELS:
        ensure_webhook_channel(signoz, name)

    return WEBHOOK_CHANNELS
