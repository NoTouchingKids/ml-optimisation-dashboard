import asyncio
import struct
import uuid
from loguru import logger
import zstandard as zstd
import msgpack
import queue
import time


class LoguruTCPSink:
    def __init__(self, host="localhost", port=9999, compression_level=3):
        self.host = host
        self.port = port
        self.compressor = zstd.ZstdCompressor(level=compression_level)
        self.queue = asyncio.Queue(1000)  # Buffer up to 1000 messages
        self.shutdown = False
        self.socket = None
        self.connected = False

    async def _connect(self):
        while not self.shutdown:
            try:
                self.socket = await asyncio.open_connection(self.host, self.port)
                self.connected = True
                print(f"Connected to log server at {self.host}:{self.port}")
                break
            except Exception as e:
                print(f"Connection error: {e}, retrying in 1 second...")
                await asyncio.sleep(1)  # Wait before retrying

    async def _send_data(self):
        while not self.shutdown:
            try:
                message = await asyncio.wait_for(self.queue.get(), timeout=0.1)
                if not message:
                    continue

                # Send message
                if self.socket:
                    writer = self.socket[1]
                    writer.write(message)
                    await writer.drain()  # Ensure data is sent before continuing
                    print("Sent message")
                    self.queue.task_done()

            except asyncio.TimeoutError:
                continue  # If no message in queue, just keep checking

            except Exception as e:
                print(f"Send error: {e}, reconnecting...")
                self.connected = False
                # Reconnect logic (this will re-establish connection)
                await self._connect()

    async def _worker(self):
        # Ensure the connection is established first
        await self._connect()

        # Start sending messages
        await asyncio.gather(self._send_data())

    async def __call__(self, message):
        # Pack and compress the message
        record = message.record
        print(record["extra"].get("client_id", "default"))

        # Extract client_id from record extra (or use a default)
        client_id = uuid.UUID(record["extra"].get("client_id", "default")).bytes

        # Create log data dict from Loguru record
        log_data = {
            "msg": record["message"],
            "time": record["time"].isoformat(),
            "levelname": record["level"].name,
            "process": record["process"].id,
            "threadName": record["thread"].name,
            "pathname": record["file"].path,
            "funcName": record["function"],
            "module": record["module"],
            "line": record["line"],
        }

        # Add all extra fields
        for key, value in record["extra"].items():
            log_data[key] = (
                str(value)
                if not isinstance(value, (str, int, float, bool, type(None)))
                else value
            )

        # Add extra fields, compress with msgpack+zstd
        packed_data = msgpack.packb(log_data)
        compressed_data = self.compressor.compress(packed_data)

        # Frame the message
        message_body = client_id + compressed_data
        framed_message = struct.pack("!I", len(message_body)) + message_body

        # Put the message into the queue (non-blocking)
        try:
            await self.queue.put(framed_message)
        except asyncio.QueueFull:
            # Handle full queue - could drop or log locally
            print("Queue is full. Message not sent.")

    async def close(self):
        self.shutdown = True
        if self.socket:
            writer = self.socket[1]
            writer.close()
            await writer.wait_closed()
        print("Connection closed.")


# Example Usage


async def main():
    # Initialize LoguruTCP sink
    log_sink = LoguruTCPSink(host="localhost", port=9999)

    # Configure a custom logger with our TCP sink
    # Note: This removes default handlers, so add them back if needed
    logger.configure(
        handlers=[
            {"sink": log_sink, "level": "DEBUG"},
        ],
        extra={"client_id": "5df52d30-6a63-482b-8802-f2e0edcb4362"},
    )
    # Create a bound logger with the client_id
    # logger.bind(client_id="5df52d30-6a63-482b-8802-f2e0edcb4362")
    # Start worker to process and send logs
    asyncio.create_task(log_sink._worker())

    # Simulate sending logs
    for i in range(5):
        logger.info(i)
        await asyncio.sleep(0.5)  # Simulate some delay between logs

    # Create a bound logger with the client_id
    # logger.bind()

    # Close the connection after sending messages
    await log_sink.close()


if __name__ == "__main__":
    asyncio.run(main())
