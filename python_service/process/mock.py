import datetime
import numpy as np
import pandas as pd
import time
from .base import BaseProcess


class MockTrainProcess(BaseProcess):
    def execute(self):
        try:
            self.log_status("started", "Starting mock training", "train")

            # Simulate some work
            time.sleep(2)

            # Log progress
            self.logger.info("Generating mock training data...")
            time.sleep(1)

            self.logger.info("Training mock model...")
            time.sleep(2)

            # Save mock model info
            # mock_model = {"trained": True, "timestamp": pd.Timestamp.now().isoformat()}

            self.log_status("completed", "Mock model trained successfully", "train")

        except Exception as e:
            self.log_status("error", f"Mock training failed: {str(e)}", "train")
            raise


class MockPredictProcess(BaseProcess):
    def execute(self):
        try:
            self.log_status("started", "Starting mock prediction", "predict")

            # Simulate work
            time.sleep(1)

            # Generate mock predictions
            num_predictions = self.config.get("forecast_periods", 30)

            # Create slightly noisy upward trend
            trend = np.linspace(0, 1, num_predictions)
            noise = np.random.normal(0, 0.1, num_predictions)
            predictions = trend + noise

            # Scale predictions to reasonable values
            predictions = 100 + (predictions * 20)

            # Generate dates
            start_date = pd.to_datetime(
                self.config.get("inference_start_date", datetime.datetime.now())
            )
            dates = pd.date_range(start=start_date, periods=num_predictions, freq="D")

            # Your long-running process logic here
            # Example: Gurobi optimization model
            for i in range(num_predictions):  # Replace with actual process
                self.logger.info(f"Optimization step {i+1}")
                time.sleep(0.2)  # Simulate work

            self.logger.info("Process completed successfully")

            results = {
                "predictions": predictions.tolist(),
                "dates": dates.strftime("%Y-%m-%d").tolist(),
                "mean": float(np.mean(predictions)),
                "std": float(np.std(predictions)),
            }

            self.log_status(
                "completed", f"Generated {num_predictions} mock predictions", "predict"
            )

            # Log results
            self.logger.info(f"Mock prediction results: {results}")

        except Exception as e:
            self.log_status("error", f"Mock prediction failed: {str(e)}", "predict")
            raise
