from abc import ABC, abstractmethod
import json
from datetime import datetime
from typing import Dict, Any
from loguru import logger
from udp_json_socket_handler import LoguruTCPSink


class BaseProcess(ABC):
    def __init__(self, client_id: str, config: Dict[str, Any]):
        self.client_id = client_id
        self.config = config
        self.log_sink = None
        self.logger = None
        # Set up logger immediately in constructor
        self.setup_logger()

    def setup_logger(self):
        # Create a custom sink for TCP streaming using our threaded implementation
        self.log_sink = LoguruTCPSink(host="localhost", port=9999)

        # Configure a custom logger with our TCP sink
        # Note: This removes default handlers, so add them back if needed
        logger.configure(
            handlers=[
                {"sink": self.log_sink, "level": "DEBUG"},
            ],
            extra={
                "client_id": self.client_id
            },  # Use the actual client_id from the instance
        )

        # Create a bound logger with the client_id
        self.logger = logger

    def log_status(self, status: str, message: str = "", process_type: str = ""):
        status_msg = {
            "status": status,
            "message": message,
            "timestamp": datetime.now().isoformat(),
            "process_type": process_type,
            "client_id": self.client_id,
        }
        # With Loguru, we can log the dict directly or as JSON
        self.logger.info(json.dumps(status_msg))

    @abstractmethod
    def execute(self) -> None:
        pass

    def cleanup(self) -> None:
        if self.log_sink:
            self.log_sink.close()
