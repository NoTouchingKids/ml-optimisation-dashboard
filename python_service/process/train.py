import numpy as np
import pandas as pd
from datetime import datetime
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import TimeSeriesSplit
from sklearn.metrics import mean_squared_error
import joblib
import json
from .base import BaseProcess


class TrainProcess(BaseProcess):
    def execute(self):
        try:
            self.log_status("started", "Starting model training", "train")

            # Extract configuration
            start_date = pd.to_datetime(self.config["start_date"])
            end_date = pd.to_datetime(self.config["end_date"])
            feature_engineering = self.config.get("feature_engineering", True)

            # Load and prepare data
            data = self._load_data(start_date, end_date)
            X, y = self._prepare_data(data, feature_engineering)

            # Train model
            model = self._train_model(X, y)

            # Save model
            model_path = f"models/{self.client_id}_model.joblib"
            joblib.dump(model, model_path)

            self.log_status(
                "completed",
                f"Model trained successfully. Saved to {model_path}",
                "train",
            )

        except Exception as e:
            self.log_status("error", f"Training failed: {str(e)}", "train")
            raise

    def _load_data(self, start_date, end_date):
        self.logger.info("Loading training data")
        # Implement your data loading logic here
        # This is a placeholder for demonstration
        dates = pd.date_range(start_date, end_date, freq="D")
        data = pd.DataFrame(
            {"date": dates, "value": np.random.normal(0, 1, len(dates))}
        )
        return data

    def _prepare_data(self, data, feature_engineering):
        self.logger.info("Preparing training data")

        if feature_engineering:
            # Add time-based features
            data["dayofweek"] = data["date"].dt.dayofweek
            data["month"] = data["date"].dt.month
            data["quarter"] = data["date"].dt.quarter

            # Add lag features
            for lag in [1, 7, 14, 30]:
                data[f"lag_{lag}"] = data["value"].shift(lag)

            # Add rolling statistics
            for window in [7, 14, 30]:
                data[f"rolling_mean_{window}"] = (
                    data["value"].rolling(window=window).mean()
                )
                data[f"rolling_std_{window}"] = (
                    data["value"].rolling(window=window).std()
                )

        # Drop NA values
        data = data.dropna()

        # Prepare features and target
        feature_cols = [col for col in data.columns if col not in ["date", "value"]]
        X = data[feature_cols]
        y = data["value"]

        # Scale features
        scaler = StandardScaler()
        X = scaler.fit_transform(X)

        return X, y

    def _train_model(self, X, y):
        self.logger.info("Training model")

        # Implement your model training logic here
        # This is a placeholder using a simple model
        from sklearn.ensemble import RandomForestRegressor

        model = RandomForestRegressor(n_estimators=100, random_state=42)

        # Time series cross-validation
        tscv = TimeSeriesSplit(n_splits=5)
        scores = []

        for train_idx, val_idx in tscv.split(X):
            X_train, X_val = X[train_idx], X[val_idx]
            y_train, y_val = y[train_idx], y[val_idx]

            model.fit(X_train, y_train)
            y_pred = model.predict(X_val)
            score = mean_squared_error(y_val, y_pred, squared=False)
            scores.append(score)

            self.logger.info(f"Fold RMSE: {score:.4f}")

        # Final fit on all data
        model.fit(X, y)

        return model
