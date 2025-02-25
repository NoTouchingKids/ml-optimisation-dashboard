import logging
from datetime import datetime
from udp_json_socket_handler import UDPJsonSocketHandler


class MLProcess:
    def __init__(self, client_id: str, config: dict):
        self.client_id = client_id
        self.config = config
        self.logger = self._setup_logger()

    def _setup_logger(self):
        logger = logging.getLogger(self.client_id)
        logger.setLevel(logging.INFO)
        handler = UDPJsonSocketHandler("localhost", 9999)
        logger.addHandler(handler)
        return logger

    def execute(self):
        """Execute the ML model process"""
        try:
            self.logger.info("Process started")

            if self.config.get("type") == "train":
                self._execute_training()
            else:
                self._execute_inference()

            self.logger.info("Process completed successfully")
        except Exception as e:
            self.logger.error(f"Process failed: {str(e)}")
        finally:
            self._cleanup()

    def _execute_training(self):
        # Implementation for model training
        self.logger.info("Training model...")
        # Add your training logic here

    def _execute_inference(self):
        # Implementation for model inference
        self.logger.info("Running inference...")
        # Add your inference logic here

    def _cleanup(self):
        """Cleanup resources"""
        for handler in self.logger.handlers[:]:
            handler.close()
            self.logger.removeHandler(handler)
