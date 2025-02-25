import signal
import sys
import os
import time
from service.process_manager import ProcessManager
from service.udp_server import UDPServer
from service.grpc_server import GRPCServer
from service.health import HealthChecker
from service.logger import setup_logging
from config import load_config


class ServiceManager:
    def __init__(self):
        self.config = load_config()
        setup_logging(self.config.log_level, "log.log")
        os.makedirs(self.config.model_path, exist_ok=True)

        self.process_manager = ProcessManager()
        self.udp_server = UDPServer(self.config.udp_host, self.config.udp_port)
        self.grpc_server = GRPCServer(
            self.config.grpc_address, self.process_manager, self.udp_server
        )
        self.health_checker = HealthChecker()
        self.running = False

    def start(self):
        try:
            self.running = True
            self.udp_server.start()
            self.grpc_server.start()
            self.health_checker.start()

            signal.signal(signal.SIGINT, self._handle_shutdown)
            signal.signal(signal.SIGTERM, self._handle_shutdown)

            print(f"Service started on {self.config.grpc_address}")

            # Windows-compatible event loop
            while self.running:
                time.sleep(1)

        except Exception as e:
            print(f"Failed to start service: {e}")
            self.stop()
            sys.exit(1)

    def _handle_shutdown(self, signum, frame):
        print("\nReceived shutdown signal. Cleaning up...")
        self.running = False
        self.stop()
        sys.exit(0)

    def stop(self):
        self.health_checker.stop()
        self.process_manager.cleanup()
        self.udp_server.stop()
        self.grpc_server.stop()


if __name__ == "__main__":
    service = ServiceManager()
    service.start()
