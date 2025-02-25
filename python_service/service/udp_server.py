import socket
import threading
import queue
from datetime import datetime
from dataclasses import dataclass
from typing import Optional


# @dataclass
# class LogMessage:
#     timestamp: int
#     client_id: str
#     message: bytes
#     process_id: int


class UDPServer:
    def __init__(self, host: str = "localhost", port: int = 9999):
        self.host = host
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.log_queue = queue.Queue()
        self.running = False

    def start(self):
        self.sock.bind((self.host, self.port))
        self.running = True
        self.receiver_thread = threading.Thread(target=self._receive_logs)
        self.receiver_thread.daemon = True
        self.receiver_thread.start()

    def stop(self):
        self.running = False
        self.sock.close()
        if self.receiver_thread:
            self.receiver_thread.join()

    def _receive_logs(self):
        while self.running:
            try:
                data, _ = self.sock.recvfrom(65535)
                if not data:
                    continue

                # Extract client ID (first 36 bytes)
                client_id = data[:36].decode("utf-8")

                message = dict(
                    timestamp=int(datetime.now().timestamp()),
                    client_id=client_id,
                    message=data,
                    process_id=client_id,
                )
                self.log_queue.put(message)

            except Exception as e:
                print(f"Error receiving log: {e}")

    def get_log(self, timeout: float = 0.1) -> Optional[dict]:
        try:
            return self.log_queue.get(timeout=timeout)
        except queue.Empty:
            return None
