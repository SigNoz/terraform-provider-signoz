"""Provide a reachable, authenticated SigNoz instance to the integration tests.

The casting provisions a root user on first boot. We log in as that root user,
mint a service-account API key (the token the Terraform provider authenticates
with), and hand back a SigNoz handle. The whole environment is created/torn down
through the --reuse / --teardown machinery in fixtures.reuse.
"""

import time
from dataclasses import dataclass

import pytest
import requests

from fixtures import foundry, reuse
from fixtures.logger import setup_logger

logger = setup_logger(__name__)

# Must match the SIGNOZ_USER_ROOT_* values in casting.yaml.
ROOT_EMAIL = "admin@integration.test"
ROOT_PASSWORD = "password123Z$"

# Service account minted for Terraform. signoz-admin so it can manage every
# resource the provider exercises.
SERVICE_ACCOUNT_NAME = "terraform-integration"
SERVICE_ACCOUNT_ROLE = "signoz-admin"


@dataclass(frozen=True)
class SigNoz:
    """A reachable, authenticated SigNoz instance."""

    endpoint: str
    access_token: str

    def __cache__(self) -> dict:
        return {"endpoint": self.endpoint, "access_token": self.access_token}

    def __log__(self) -> str:
        return f"signoz(endpoint={self.endpoint})"


def _login_as_root(endpoint: str, email: str, password: str, *, ready_timeout: float = 240.0) -> str:
    """Return a bearer access token for the root user.

    Retries until SigNoz is up and the root user has been reconciled (it is
    created asynchronously shortly after the container starts).
    """
    deadline = time.time() + ready_timeout
    last = None

    while time.time() < deadline:
        try:
            ctx = requests.get(
                f"{endpoint}/api/v2/sessions/context",
                params={"email": email, "ref": endpoint},
                timeout=10,
            )
            if ctx.status_code == 200 and ctx.json().get("data", {}).get("orgs"):
                org_id = ctx.json()["data"]["orgs"][0]["id"]

                login = requests.post(
                    f"{endpoint}/api/v2/sessions/email_password",
                    json={"email": email, "password": password, "orgId": org_id},
                    timeout=10,
                )
                if login.status_code == 200:
                    logger.info("logged in as root user %s", email)
                    return login.json()["data"]["accessToken"]

                last = (login.status_code, login.text[:200])
            else:
                last = (ctx.status_code, ctx.text[:200])
        except requests.RequestException as err:
            last = err

        time.sleep(3)

    raise TimeoutError(f"could not log in as {email} within {ready_timeout}s (last={last})")


def apply_license(endpoint: str, bearer_token: str, license_key: str) -> None:
    """Apply a license key to a freshly started SigNoz via the admin API.

    No-op when no key is given, so community-only runs (and forks without the
    secret) still work.
    """
    if not license_key:
        logger.info("no license key provided; skipping license application")
        return

    resp = requests.post(
        f"{endpoint}/api/v3/licenses",
        json={"key": license_key},
        headers={"Authorization": f"Bearer {bearer_token}"},
        timeout=30,
    )
    assert resp.status_code == 202, resp.text

    logger.info("applied SigNoz license")


def mint_service_account_key(
    endpoint: str,
    bearer_token: str,
    *,
    name: str = SERVICE_ACCOUNT_NAME,
    role: str = SERVICE_ACCOUNT_ROLE,
) -> str:
    """Create a service account, assign it a role, and return a fresh API key."""
    sa = requests.post(
        f"{endpoint}/api/v1/service_accounts",
        json={"name": name},
        headers={"Authorization": f"Bearer {bearer_token}"},
        timeout=10,
    )
    assert sa.status_code == 201, sa.text
    sa_id = sa.json()["data"]["id"]

    roles = requests.get(
        f"{endpoint}/api/v1/roles",
        headers={"Authorization": f"Bearer {bearer_token}"},
        timeout=10,
    )
    assert roles.status_code == 200, roles.text
    role_id = next(r["id"] for r in roles.json()["data"] if r["name"] == role)

    assign = requests.post(
        f"{endpoint}/api/v1/service_accounts/{sa_id}/roles",
        json={"id": role_id},
        headers={"Authorization": f"Bearer {bearer_token}"},
        timeout=10,
    )
    assert assign.status_code == 204, assign.text

    key = requests.post(
        f"{endpoint}/api/v1/service_accounts/{sa_id}/keys",
        json={"name": "terraform-integration", "expiresAt": 0},
        headers={"Authorization": f"Bearer {bearer_token}"},
        timeout=10,
    )
    assert key.status_code == 201, key.text

    logger.info("minted service-account key for %s (role %s)", name, role)
    return key.json()["data"]["key"]


@pytest.fixture(scope="session")
def signoz(request: pytest.FixtureRequest, pytestconfig: pytest.Config) -> SigNoz:
    """A SigNoz instance with a service-account access token for Terraform."""
    foundryctl = request.config.getoption("--foundry-binary-path")

    def empty() -> SigNoz:
        return SigNoz(endpoint="", access_token="")

    def create() -> SigNoz:
        endpoint = foundry.cast(foundryctl)
        bearer_token = _login_as_root(endpoint, ROOT_EMAIL, ROOT_PASSWORD)

        apply_license(endpoint, bearer_token, request.config.getoption("--license-key"))

        access_token = mint_service_account_key(endpoint, bearer_token)
        return SigNoz(endpoint=endpoint, access_token=access_token)

    def delete(_: SigNoz) -> None:
        foundry.teardown()

    def restore(cache: dict) -> SigNoz:
        return SigNoz(endpoint=cache["endpoint"], access_token=cache["access_token"])

    return reuse.wrap(request, pytestconfig, "signoz", empty, create, delete, restore)
