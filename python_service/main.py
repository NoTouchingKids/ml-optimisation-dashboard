import logging
import signal
import sys
import os
import asyncio
from service.process_manager import ProcessManager
from service.udp_server import TCPLogServer
from service.grpc_server import GRPCServer
from service.health import HealthChecker
from service.logger import setup_logging
from config import load_config
import win32api


class ServiceManager:
    def __init__(self):
        self.config = load_config()
        setup_logging(self.config.log_level, "log.log")
        os.makedirs(self.config.model_path, exist_ok=True)
        self.logger = logging.getLogger(__file__)
        self.process_manager = ProcessManager()
        self.udp_server = TCPLogServer(self.config.udp_host, self.config.udp_port)
        self.grpc_server = GRPCServer(
            self.config.grpc_address, self.process_manager, self.udp_server
        )
        self.health_checker = HealthChecker()
        self.running = False

    async def start(self):
        self.running = True
        self.logger.info("Service starting...")
        loop = asyncio.get_running_loop()

        # Register the Win32 console control handler on Windows
        if sys.platform.startswith("win"):

            def console_ctrl_handler(ctrl_type):
                # Schedule the shutdown asynchronously
                loop.call_soon_threadsafe(lambda: asyncio.create_task(self.shutdown()))
                # Return True to indicate that the event has been handled.
                return True

            win32api.SetConsoleCtrlHandler(console_ctrl_handler, True)
            self.logger.info("Console control handler registered (Windows).")
        else:
            # For non-Windows platforms, you can use add_signal_handler
            for sig in (signal.SIGINT, signal.SIGTERM):
                try:
                    loop.add_signal_handler(
                        sig, lambda: asyncio.create_task(self.shutdown())
                    )
                except NotImplementedError:
                    self.logger.warning(
                        f"Signal handler for {sig} not supported; skipping."
                    )

        # Start UDP server concurrently
        asyncio.create_task(self.udp_server.start())
        self.logger.info("UDP Server started.")

        # Start health checker concurrently (if it's async)
        asyncio.create_task(asyncio.to_thread(self.health_checker.start))
        self.logger.info("Health Checker started.")

        # Start gRPC server and wait for its termination
        await self.grpc_server.start()
        self.logger.info("gRPC Server started.")

        print(f"Service started on {self.config.grpc_address}")
        await self.grpc_server.server.wait_for_termination()

    async def shutdown(self):
        print("\nReceived shutdown signal. Cleaning up...")
        self.running = False
        await self.stop()
        # sys.exit(0)

    async def stop(self):
        self.health_checker.stop()
        self.process_manager.cleanup()
        await self.udp_server.stop()
        await self.grpc_server.stop()


if __name__ == "__main__":

    if sys.platform.startswith("win"):
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    service = ServiceManager()
    try:
        asyncio.run(service.start())
    except KeyboardInterrupt:
        pass
