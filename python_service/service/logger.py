import logging
import sys
from typing import Optional


DETAILED_LOG_FORMAT = (
    "%(asctime)s [%(levelname)s] [%(name)s] "
    "[%(module)s:%(funcName)s:%(lineno)d] "
    "[Thread:%(threadName)s] [Process:%(process)d] "
    "%(message)s"
)

DETAILED_LOG_FORMAT_V2 = (
    "%(asctime)s [%(levelname)s] "
    "[%(module)s:%(funcName)s:%(lineno)d] "
    "[Thread:%(threadName)s] [Process:%(process)d] "
    "%(message)s"
)


def setup_logging(level: str = "INFO", log_file: Optional[str] = None):
    logging.basicConfig(
        level=logging.DEBUG,
        # level=getattr(logging, level.upper()),
        format=DETAILED_LOG_FORMAT_V2,
        datefmt="%Y-%m-%d %H:%M:%S",
        handlers=[
            logging.StreamHandler(sys.stdout),
            *([] if not log_file else [logging.FileHandler(log_file)]),
        ],
    )

    # Suppress noisy logs
    # logging.getLogger("matplotlib").setLevel(logging.WARNING)
    # logging.getLogger("PIL").setLevel(logging.WARNING)
