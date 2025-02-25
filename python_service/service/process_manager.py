import multiprocessing
from typing import Dict, Any
from process.mock import MockPredictProcess


def run_process(client_id: str, config: Dict[str, Any]):
    try:
        # process_type = config.get("type", "train")
        process = MockPredictProcess(client_id, config)
        process.execute()
    except Exception as e:
        print(f"Process failed: {e}")
    finally:
        if hasattr(process, "cleanup"):
            process.cleanup()


class ProcessManager:
    def __init__(self):
        self.processes: Dict[str, multiprocessing.Process] = {}

    def start_process(
        self, client_id: str, config: Dict[str, Any]
    ) -> multiprocessing.Process:
        process = multiprocessing.Process(target=run_process, args=(client_id, config))
        process.daemon = True
        process.start()

        self.processes[client_id] = process
        return process

    def stop_process(self, client_id: str) -> bool:
        process = self.processes.get(client_id)
        if process and process.is_alive():
            process.terminate()
            process.join(timeout=5)
            del self.processes[client_id]
            return True
        return False

    def cleanup(self):
        for process in self.processes.values():
            if process.is_alive():
                process.terminate()
                process.join(timeout=5)
