from process.base import BaseProcess
import pandas as pd
import numpy as np
from prophet import Prophet

# import matplotlib.pyplot as plt
# from io import BytesIO
# import base64
import json
import time


class TimeSeriesForecaster(BaseProcess):
    def __init__(self, client_id: str, config: dict):
        super().__init__(client_id, config)
        self.model = None

    def execute(self):
        try:
            self.log_status("started", "Starting time series forecasting", "forecast")

            # Get data from config or generate sample data
            data = self.config.get("data")
            if not data:
                self.logger.info("No data provided, generating synthetic data")
                # Generate synthetic time series data with trend and seasonality
                np.random.seed(42)
                periods = 365 * 2  # 2 years of daily data
                dates = pd.date_range(end=pd.Timestamp.now(), periods=periods, freq="D")

                # Create time features
                t = np.arange(periods)
                # Trend + weekly seasonality + annual seasonality + noise
                values = (
                    100
                    + 0.05 * t
                    + 10 * np.sin(2 * np.pi * t / 7)
                    + 20 * np.sin(2 * np.pi * t / 365)
                    + np.random.normal(0, 5, periods)
                )

                data = pd.DataFrame({"ds": dates, "y": values})
                self.logger.info(f"Generated {periods} data points")
            elif isinstance(data, list):
                # Convert list to dataframe
                end_date = pd.Timestamp.now()
                dates = pd.date_range(end=end_date, periods=len(data), freq="D")
                data = pd.DataFrame({"ds": dates, "y": data})

            # Configure model parameters
            seasonality_mode = self.config.get("seasonality_mode", "additive")
            changepoint_prior_scale = self.config.get("changepoint_prior_scale", 0.05)
            self.logger.info(
                f"Model configuration: seasonality_mode={seasonality_mode}, changepoint_prior_scale={changepoint_prior_scale}"
            )

            # Create and train Prophet model
            self.logger.info("Training forecasting model")
            start_time = time.time()

            model = Prophet(
                seasonality_mode=seasonality_mode,
                changepoint_prior_scale=changepoint_prior_scale,
                daily_seasonality=self.config.get("daily_seasonality", "auto"),
                weekly_seasonality=self.config.get("weekly_seasonality", True),
                yearly_seasonality=self.config.get("yearly_seasonality", True),
            )

            # Add additional seasonality if specified
            if self.config.get("add_monthly_seasonality", False):
                model.add_seasonality(name="monthly", period=30.5, fourier_order=5)

            # Fit model
            model.fit(data)
            self.model = model
            train_time = time.time() - start_time
            self.logger.info(f"Model training completed in {train_time:.2f} seconds")

            # Generate forecast
            forecast_periods = self.config.get("forecast_periods", 30)
            self.logger.info(f"Generating forecast for {forecast_periods} periods")

            future = model.make_future_dataframe(periods=forecast_periods)
            forecast = model.predict(future)

            # Generate plots
            # plots = {}
            # if self.config.get("generate_plots", True):
            #     self.logger.info("Generating forecast plots")
            #     # Forecast plot
            #     fig1 = model.plot(forecast)
            #     buf1 = BytesIO()
            #     fig1.savefig(buf1, format="png", dpi=100)
            #     plots["forecast"] = base64.b64encode(buf1.getvalue()).decode("utf-8")
            #     plt.close(fig1)

            #     # Components plot
            #     fig2 = model.plot_components(forecast)
            #     buf2 = BytesIO()
            #     fig2.savefig(buf2, format="png", dpi=100)
            #     plots["components"] = base64.b64encode(buf2.getvalue()).decode("utf-8")
            #     plt.close(fig2)

            # Extract and log forecast metrics
            forecast_result = forecast[["ds", "yhat", "yhat_lower", "yhat_upper"]].tail(
                forecast_periods
            )
            metrics = {
                "mean_forecast": float(forecast_result["yhat"].mean()),
                "min_forecast": float(forecast_result["yhat"].min()),
                "max_forecast": float(forecast_result["yhat"].max()),
                "forecast_range": float(
                    forecast_result["yhat"].max() - forecast_result["yhat"].min()
                ),
            }

            self.logger.info(
                json.dumps(
                    {
                        "forecast_metrics": metrics,
                        "train_time_seconds": train_time,
                        "data_points": len(data),
                        "forecast_periods": forecast_periods,
                    }
                )
            )

            # Prepare and return results
            results = {
                "forecast": forecast_result.to_dict("records"),
                "metrics": metrics,
                # "plots": plots if plots else None,
                "training_info": {
                    "data_points": len(data),
                    "train_time_seconds": train_time,
                },
            }

            self.log_status(
                "completed",
                f"Forecast generated for {forecast_periods} periods",
                "forecast",
            )
            return results

        except Exception as e:
            self.logger.error(f"Error in time series forecasting: {str(e)}")
            self.log_status("error", f"Forecasting failed: {str(e)}", "forecast")
            raise
