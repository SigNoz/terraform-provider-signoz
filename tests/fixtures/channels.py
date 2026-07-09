"""Seed the notification channels the example alerts and route policies reference.

SigNoz validates that a referenced channel exists, so every channel named in an example must be created
before the CRUD cycle. Names are discovered from the example .tf files (plus a small base set the
regression tests use) so new examples are covered automatically.
"""

import re

import pytest
import requests

from fixtures.logger import setup_logger
from fixtures.signoz import SigNoz
from fixtures.terraform import EXAMPLES

logger = setup_logger(__name__)

WEBHOOK_URL = "https://example.com/webhook"
# Always ensured, independent of the examples (used by the regression tests).
BASE_CHANNELS = ("slack", "pagerduty")

_TIMEOUT = 10
_CHANNELS_RE = re.compile(r"channels\s*=\s*\[([^\]]*)\]")


def _discover_channel_names() -> tuple[str, ...]:
    """Return every channel name referenced by a `channels = [...]` block in the examples."""
    names: set[str] = set(BASE_CHANNELS)
    for tf_file in EXAMPLES.rglob("*.tf"):
        for match in _CHANNELS_RE.findall(tf_file.read_text()):
            for raw in match.split(","):
                name = raw.strip().strip('"')
                if name:
                    names.add(name)
    return tuple(sorted(names))


def ensure_webhook_channel(signoz: SigNoz, name: str) -> None:
    """Create a webhook notification channel by name if it does not already exist."""
    headers = {"SIGNOZ-API-KEY": signoz.access_token}

    listing = requests.get(f"{signoz.endpoint}/api/v1/channels", headers=headers, timeout=_TIMEOUT)
    listing.raise_for_status()
    existing = {channel.get("name") for channel in (listing.json().get("data") or [])}
    if name in existing:
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
    """Ensure every channel referenced by the examples (and the base set) exists."""
    names = _discover_channel_names()
    for name in names:
        ensure_webhook_channel(signoz, name)

    return names
