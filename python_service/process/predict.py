import numpy as np
import pandas as pd
from datetime import datetime
import joblib
import json
from .base import BaseProcess


class PredictProcess(BaseProcess):
    def execute(self):
        try:
            self.log_status("started", "Starting prediction", "predict")

            # Load model
            model_path = f"models/{self.client_id}_model.joblib"
            model = joblib.load(model_path)

            # Prepare data
            data = self._prepare_data()

            # Generate predictions
            predictions = model.predict(data)

            # Format results
            results = {
                "predictions": predictions.tolist(),
                "dates": self._generate_forecast_dates(),
                "status": "completed",
            }

            self.log_status("completed", json.dumps(results), "predict")

        except Exception as e:
            self.log_status("error", f"Prediction failed: {str(e)}", "predict")
            raise

    def _prepare_data(self):
        self.logger.info("Preparing prediction data")

        # Get input data
        input_data = np.array(self.config["data"])
        feature_engineering = self.config.get("feature_engineering", True)

        if feature_engineering:
            # Convert to DataFrame
            dates = pd.date_range(
                start=self.config["inference_start_date"],
                periods=len(input_data),
                freq="D",
            )
            data = pd.DataFrame({"date": dates, "value": input_data})

            # Add features similar to training
            features = []

            # Time-based features
            features.extend(
                [
                    data["date"].dt.dayofweek,
                    data["date"].dt.month,
                    data["date"].dt.quarter,
                ]
            )

            # Lag features
            for lag in [1, 7, 14, 30]:
                features.append(data["value"].shift(lag))

            # Rolling statistics
            for window in [7, 14, 30]:
                features.extend(
                    [
                        data["value"].rolling(window=window).mean(),
                        data["value"].rolling(window=window).std(),
                    ]
                )

            # Combine features
            feature_matrix = np.column_stack(features)

            # Handle missing values
            feature_matrix = np.nan_to_num(feature_matrix, 0)

        else:
            feature_matrix = input_data.reshape(-1, 1)

        return feature_matrix

    def _generate_forecast_dates(self):
        forecast_periods = self.config.get("forecast_periods", 30)
        start_date = pd.to_datetime(self.config["inference_start_date"])
        dates = pd.date_range(start=start_date, periods=forecast_periods, freq="D")
        return dates.strftime("%Y-%m-%d").tolist()
