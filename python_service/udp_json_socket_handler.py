from datetime import datetime
import logging
import socket
from logging.handlers import SocketHandler
import msgpack
import zstandard as zstd
from uuid import uuid4

# Constants
UDP_IP = "127.0.0.1"
UDP_PORT = 5005
UUID_LENGTH = 36  # Standard UUID string length
UUID = uuid4()


class UDPJsonSocketHandler(SocketHandler):
    """
    A UDP socket handler that prefixes client_id to base64 encoded JSON data.
    Format: [36-byte UUID][base64 encoded json]
    """

    def __init__(self, host: str = "localhost", port: int = 9999) -> None:
        super().__init__(host, port)
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.compressor = zstd.ZstdCompressor()

    def makeSocket(self) -> None:
        pass  # UDP doesn't need connection

    def makePickle(self, record: logging.LogRecord) -> bytes:
        """
        Creates a binary message with format:
        [36-byte UUID][base64 encoded json]
        """
        # Extract client_id from record name (assuming it's a UUID)
        client_id = record.name.encode("utf-8")

        # Prepare log data
        log_data = {
            "msg": record.getMessage(),
            "time": datetime.fromtimestamp(record.created).isoformat(),
            "levelname": record.levelname,
            "process": record.process,
            "thread": record.thread,
            "threadName": record.threadName,
        }

        # Add any extra attributes
        for key, value in record.__dict__.items():
            if key not in logging.LogRecord.__dict__ and key not in [
                "args",
                "exc_info",
                "exc_text",
                "stack_info",
                "created",
                "msecs",
                "relativeCreated",
                "msg",
                "name",
            ]:
                log_data[key] = str(value)

        # Add status if completion/error message
        if "Process completed successfully" in log_data["msg"]:
            log_data["status"] = "completed"
        elif "Process failed" in log_data["msg"]:
            log_data["status"] = "error"

        # # Encode JSON to base64
        # json_bytes = json.dumps(log_data).encode("utf-8")
        # b64_data = base64.b64encode(json_bytes)

        packed_data = msgpack.packb(log_data)
        compressed_data = self.compressor.compress(packed_data)

        # Combine client_id and base64 data
        return client_id + compressed_data

    def send(self, data: bytes) -> None:
        """Send the encoded log record via UDP."""
        try:
            self.sock.sendto(data, (self.host, self.port))
        except Exception:
            self.handleError(None)
