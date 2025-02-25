from abc import ABC, abstractmethod
import logging
import json
from datetime import datetime
from typing import Dict, Any


class BaseProcess(ABC):
    def __init__(self, client_id: str, config: Dict[str, Any]):
        self.client_id = client_id
        self.config = config
        self.logger = self._setup_logger()

    def _setup_logger(self) -> logging.Logger:
        logger = logging.getLogger(self.client_id)
        logger.setLevel(logging.DEBUG)

        # Add UDP handler for logs
        from udp_json_socket_handler import UDPJsonSocketHandler

        handler = UDPJsonSocketHandler("localhost", 9999)
        logger.addHandler(handler)

        return logger

    def log_status(self, status: str, message: str = "", process_type: str = ""):
        status_msg = {
            "status": status,
            "message": message,
            "timestamp": datetime.utcnow().isoformat(),
            "process_type": process_type,
            "client_id": self.client_id,
        }
        self.logger.info(json.dumps(status_msg))

    @abstractmethod
    def execute(self) -> None:
        pass

    def cleanup(self) -> None:
        handlers = self.logger.handlers[:]
        for handler in handlers:
            handler.close()
            self.logger.removeHandler(handler)
