import threading
import time
from typing import Dict, Any
import psutil
import logging


class HealthChecker:
    def __init__(self, check_interval: int = 60):
        self.check_interval = check_interval
        self.is_running = False
        self._thread: threading.Thread = None
        self.last_check: Dict[str, Any] = {}

    def start(self):
        self.is_running = True
        self._thread = threading.Thread(target=self._run_checks, daemon=True)
        self._thread.start()

    def stop(self):
        self.is_running = False
        if self._thread:
            self._thread.join()

    def _run_checks(self):
        while self.is_running:
            try:
                self._check_system_health()
                time.sleep(self.check_interval)
            except Exception as e:
                logging.error(f"Health check failed: {e}")

    def _check_system_health(self):
        cpu_percent = psutil.cpu_percent(interval=1)
        memory = psutil.virtual_memory()
        disk = psutil.disk_usage("/")

        self.last_check = {
            "timestamp": time.time(),
            "cpu_usage": cpu_percent,
            "memory_usage": memory.percent,
            "disk_usage": disk.percent,
            "status": (
                "healthy"
                if all([cpu_percent < 90, memory.percent < 90, disk.percent < 90])
                else "warning"
            ),
        }

        if self.last_check["status"] == "warning":
            logging.warning(f"System resources high: {self.last_check}")
