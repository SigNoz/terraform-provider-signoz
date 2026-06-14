import logging


def setup_logger(name: str) -> logging.Logger:
    """Return a named logger; output is driven by pytest's log_cli config."""
    return logging.getLogger(name)
