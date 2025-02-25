import json
import asyncpg
import numpy as np
import pandas as pd
from datetime import datetime, timedelta
import asyncio
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TimeSeriesDataGenerator:
    def __init__(self):
        self.db_config = {
            "user": "postgres",
            "password": "postgres",
            "host": "localhost",
            "port": "5432",
            "database": "logsdb",
        }

    async def create_tables(self, pool):
        """Create the necessary tables and hypertable"""
        async with pool.acquire() as conn:
            # Create the training data table
            await conn.execute(
                """
                CREATE TABLE IF NOT EXISTS training_data (
                    timestamp TIMESTAMPTZ NOT NULL,
                    value DOUBLE PRECISION NOT NULL,
                    series_id TEXT NOT NULL,
                    metadata JSONB
                );
            """
            )

            # Create hypertable
            try:
                await conn.execute(
                    """
                    SELECT create_hypertable('training_data', 'timestamp', 
                        if_not_exists => TRUE);
                """
                )
            except Exception as e:
                logger.warning(f"Hypertable might already exist: {e}")

            # Create index on series_id
            await conn.execute(
                """
                CREATE INDEX IF NOT EXISTS idx_training_data_series 
                ON training_data (series_id, timestamp DESC);
            """
            )

    def generate_synthetic_data(self, start_date, num_days, num_series=5):
        """Generate synthetic time series data"""
        dates = pd.date_range(start=start_date, periods=num_days, freq="D")
        all_data = []

        for series_id in range(num_series):
            # Base trend
            trend = np.linspace(0, 2, num_days)

            # Seasonal component (yearly)
            seasonal = 10 * np.sin(2 * np.pi * np.arange(num_days) / 365)

            # Weekly pattern
            weekly = 5 * np.sin(2 * np.pi * np.arange(num_days) / 7)

            # Random noise
            noise = np.random.normal(0, 1, num_days)

            # Combine components
            values = 100 + trend * 10 + seasonal + weekly + noise

            # Create metadata
            metadata = {
                "frequency": "daily",
                "pattern": "synthetic",
                "components": ["trend", "seasonal", "weekly", "noise"],
                "base_value": 100,
            }

            # Combine into series
            for date, value in zip(dates, values):
                all_data.append(
                    {
                        "timestamp": date.to_pydatetime(),
                        "value": float(value),
                        "series_id": f"series_{series_id}",
                        "metadata": json.dumps(metadata),
                    }
                )

        return all_data

    async def insert_data(self, pool, data):
        """Insert data into TimescaleDB"""
        async with pool.acquire() as conn:
            # Prepare the insert statement
            stmt = await conn.prepare(
                """
                INSERT INTO training_data (timestamp, value, series_id, metadata)
                VALUES ($1, $2, $3, $4)
            """
            )

            # Insert in chunks
            for i in range(0, len(data)):
                row = data[i]
                await stmt.fetch(
                    row["timestamp"],
                    row["value"],
                    row["series_id"],
                    row["metadata"],
                )
                # chunk = data[i : i + chunk_size]
                # await asyncio.gather(
                #     *[
                #         stmt.fetch(
                #             row["timestamp"],
                #             row["value"],
                #             row["series_id"],
                #             row["metadata"],
                #         )
                #         async for row in chunk
                #     ]
                # )

    async def setup_and_generate(self):
        """Main function to set up database and generate data"""
        try:
            # Create connection pool
            pool = await asyncpg.create_pool(**self.db_config)

            # Create tables
            await self.create_tables(pool)

            # Generate synthetic data
            start_date = datetime(2023, 1, 1)
            num_days = 365 * 2  # 2 years of daily data
            data = self.generate_synthetic_data(start_date, num_days)

            # Insert data
            logger.info("Inserting data into TimescaleDB...")
            await self.insert_data(pool, data)
            logger.info("Data insertion completed!")

            # Close pool
            await pool.close()

        except Exception as e:
            logger.error(f"Error: {e}")
            raise


async def main():
    generator = TimeSeriesDataGenerator()
    await generator.setup_and_generate()


if __name__ == "__main__":
    asyncio.run(main())
