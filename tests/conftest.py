import pytest

pytest_plugins = [
    "fixtures.foundry",
    "fixtures.signoz",
    "fixtures.terraform",
]


def pytest_addoption(parser: pytest.Parser):
    parser.addoption(
        "--keep-env",
        action="store_true",
        default=False,
        help="Leave the SigNoz environment running after the run (skip foundry teardown).",
    )
