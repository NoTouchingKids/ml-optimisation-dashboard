import asyncio
import socket
import queue
import struct
import threading
from time import sleep
import time
import msgpack
import zstandard as zstd

# from uuid import uuid4
import uuid

# Constants
UDP_IP = "127.0.0.1"
UDP_PORT = 5005
UUID_LENGTH = 16  # Standard UUID string length
# UUID = uuid4()


# class UDPJsonSocketHandler(SocketHandler):
#     """
#     A UDP socket handler that prefixes client_id to base64 encoded JSON data.
#     Format: [36-byte UUID][base64 encoded json]
#     """

#     def __init__(self, host: str = "localhost", port: int = 9999) -> None:
#         super().__init__(host, port)
#         self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
#         self.compressor = zstd.ZstdCompressor()

#     def makeSocket(self) -> None:
#         pass  # UDP doesn't need connection

#     def makePickle(self, record: logging.LogRecord) -> bytes:
#         """
#         Creates a binary message with format:
#         [36-byte UUID][base64 encoded json]
#         """
#         # Extract client_id from record name (assuming it's a UUID)
#         client_id = record.name.encode("utf-8")

#         # Prepare log data
#         log_data = {
#             "msg": record.getMessage(),
#             "time": datetime.fromtimestamp(record.created).isoformat(),
#             "levelname": record.levelname,
#             "process": record.process,
#             "thread": record.thread,
#             "threadName": record.threadName,
#         }

#         # Add any extra attributes
#         for key, value in record.__dict__.items():
#             if key not in logging.LogRecord.__dict__ and key not in [
#                 "args",
#                 "exc_info",
#                 "exc_text",
#                 "stack_info",
#                 "created",
#                 "msecs",
#                 "relativeCreated",
#                 "msg",
#                 "name",
#             ]:
#                 log_data[key] = str(value)

#         # Add status if completion/error message
#         if "Process completed successfully" in log_data["msg"]:
#             log_data["status"] = "completed"
#         elif "Process failed" in log_data["msg"]:
#             log_data["status"] = "error"

#         # # Encode JSON to base64
#         # json_bytes = json.dumps(log_data).encode("utf-8")
#         # b64_data = base64.b64encode(json_bytes)

#         packed_data = msgpack.packb(log_data)
#         compressed_data = self.compressor.compress(packed_data)

#         # Combine client_id and base64 data
#         return client_id + compressed_data

#     def send(self, data: bytes) -> None:
#         """Send the encoded log record via UDP."""
#         try:
#             self.sock.sendto(data, (self.host, self.port))
#         except Exception:
#             self.handleError(None)


# class LoguruTCPSink:
#     def __init__(self, host="localhost", port=9999, compression_level=3):
#         self.host = host
#         self.port = port
#         self.compressor = zstd.ZstdCompressor(level=compression_level)
#         self.queue = queue.Queue(1000)  # Buffer up to 1000 messages
#         self.shutdown = threading.Event()
#         self.socket = None
#         self.connected = False

#         # Start worker thread
#         self.worker = threading.Thread(target=self._worker_thread, daemon=True)
#         self.worker.start()

#     def _worker_thread(self):
#         while not self.shutdown.is_set():
#             # Try to establish connection if needed
#             if not self.connected:
#                 try:
#                     self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#                     self.socket.connect((self.host, self.port))
#                     self.connected = True
#                     print(f"Connected to log server at {self.host}:{self.port}")
#                 except socket.error as e:
#                     print(f"Connection error: {e}, retrying in 1 second...")
#                     # if self.socket:
#                     #     try:
#                     #         print(f"Closing Connection: {e}, Closing in 1 second...")
#                     #         self.socket.close()
#                     #     except:
#                     #         pass
#                     sleep(0.1)  # Wait before retry
#                     continue

#             # Process log messages from queue
#             try:
#                 print("Sending data")
#                 # Non-blocking get with timeout
#                 message = self.queue.get(timeout=0.001)
#                 self.socket.sendall(message)
#                 self.queue.task_done()  # Mark as processed
#             except queue.Empty:
#                 # Just continue the loop if no messages
#                 continue
#             except socket.error as e:
#                 print(f"Send error: {e}, reconnecting...")
#                 self.connected = False
#                 # Don't close the socket here, we'll create a new one on reconnect

#                 # Put the message back in the queue
#                 try:
#                     self.queue.put(message)
#                 except Exception as e:
#                     print(e)
#                 except queue.Full:
#                     # Queue might be full
#                     continue
#                     # Just continue the loop if no messages
#             sleep(0.1)
#             # Small delay before retry

#     def __call__(self, message):
#         # Pack and compress the message
#         record = message.record

#         # Extract client_id from record extra (or use a default)
#         client_id = uuid.UUID(record["extra"].get("client_id", "default")).bytes

#         # Create log data dict from Loguru record
#         log_data = {
#             "msg": record["message"],
#             "time": record["time"].isoformat(),
#             "levelname": record["level"].name,
#             "process": record["process"].id,
#             "threadName": record["thread"].name,
#             "pathname": record["file"].path,
#             "funcName": record["function"],
#             "module": record["module"],
#             "line": record["line"],
#         }

#         # Add all extra fields
#         for key, value in record["extra"].items():
#             log_data[key] = (
#                 str(value)
#                 if not isinstance(value, (str, int, float, bool, type(None)))
#                 else value
#             )

#         # Add extra fields, compress with msgpack+zstd
#         packed_data = msgpack.packb(log_data)
#         compressed_data = self.compressor.compress(packed_data)

#         # Frame the message
#         message_body = client_id + compressed_data
#         framed_message = struct.pack("!I", len(message_body)) + message_body

#         # Queue for sending (non-blocking)
#         try:
#             self.queue.put_nowait(framed_message)
#         except queue.Full:
#             # Handle full queue - could drop or log locally
#             pass

#     def close(self):
#         self.shutdown.set()
#         if self.worker.is_alive():
#             self.worker.join(timeout=2.0)


# class LoguruTCPSink:
#     def __init__(self, host="localhost", port=9999, compression_level=3):
#         self.host = host
#         self.port = port
#         self.compressor = zstd.ZstdCompressor(level=compression_level)
#         self.queue = asyncio.Queue(1000)  # Buffer up to 1000 messages
#         self.shutdown = False
#         self.socket = None
#         self.connected = False

#     async def _connect(self):
#         while not self.shutdown:
#             try:
#                 self.socket = await asyncio.open_connection(self.host, self.port)
#                 self.connected = True
#                 print(f"Connected to log server at {self.host}:{self.port}")
#                 break
#             except Exception as e:
#                 print(f"Connection error: {e}, retrying in 1 second...")
#                 await asyncio.sleep(1)  # Wait before retrying

#     async def _send_data(self):
#         while not self.shutdown:
#             try:
#                 message = await asyncio.wait_for(self.queue.get(), timeout=0.1)
#                 if not message:
#                     continue

#                 # Send message
#                 if self.socket:
#                     writer = self.socket[1]
#                     writer.write(message)
#                     await writer.drain()  # Ensure data is sent before continuing
#                     print("Sent message")
#                     self.queue.task_done()

#             except asyncio.TimeoutError:
#                 continue  # If no message in queue, just keep checking

#             except Exception as e:
#                 print(f"Send error: {e}, reconnecting...")
#                 self.connected = False
#                 # Reconnect logic (this will re-establish connection)
#                 await self._connect()

#     async def _worker(self):
#         # Ensure the connection is established first
#         await self._connect()

#         # Start sending messages
#         await asyncio.gather(self._send_data())

#     async def __call__(self, message):
#         # Pack and compress the message
#         record = message.record
#         print(record["extra"].get("client_id", "default"))

#         # Extract client_id from record extra (or use a default)
#         client_id = uuid.UUID(record["extra"].get("client_id", "default")).bytes

#         # Create log data dict from Loguru record
#         log_data = {
#             "msg": record["message"],
#             "time": record["time"].isoformat(),
#             "levelname": record["level"].name,
#             "process": record["process"].id,
#             "threadName": record["thread"].name,
#             "pathname": record["file"].path,
#             "funcName": record["function"],
#             "module": record["module"],
#             "line": record["line"],
#         }

#         # Add all extra fields
#         for key, value in record["extra"].items():
#             log_data[key] = (
#                 str(value)
#                 if not isinstance(value, (str, int, float, bool, type(None)))
#                 else value
#             )

#         # Add extra fields, compress with msgpack+zstd
#         packed_data = msgpack.packb(log_data)
#         compressed_data = self.compressor.compress(packed_data)

#         # Frame the message
#         message_body = client_id + compressed_data
#         framed_message = struct.pack("!I", len(message_body)) + message_body

#         # Put the message into the queue (non-blocking)
#         try:
#             await self.queue.put(framed_message)
#         except asyncio.QueueFull:
#             # Handle full queue - could drop or log locally
#             print("Queue is full. Message not sent.")

#     async def close(self):
#         self.shutdown = True
#         if self.socket:
#             writer = self.socket[1]
#             writer.close()
#             await writer.wait_closed()
#         print("Connection closed.")


class LoguruTCPSink:
    def __init__(self, host="localhost", port=9999, compression_level=3):
        self.host = host
        self.port = port
        self.compressor = zstd.ZstdCompressor(level=compression_level)
        self.queue = queue.Queue(1000)  # Buffer up to 1000 messages
        self.shutdown = False
        self.socket = None
        self.connected = False
        self.lock = threading.Lock()

        # Start the worker thread
        self.worker_thread = threading.Thread(target=self._worker, daemon=True)
        self.worker_thread.start()

    def _connect(self):
        while not self.shutdown:
            try:
                with self.lock:
                    self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                    self.socket.connect((self.host, self.port))
                    self.connected = True
                print(f"Connected to log server at {self.host}:{self.port}")
                break
            except Exception as e:
                print(f"Connection error: {e}, retrying in 1 second...")
                time.sleep(1)  # Wait before retrying

    def _send_data(self):
        while not self.shutdown:
            try:
                try:
                    # Get with timeout to allow for shutdown checks
                    message = self.queue.get(timeout=0.1)
                except queue.Empty:
                    continue

                # Send message
                with self.lock:
                    if self.socket and self.connected:
                        self.socket.sendall(message)
                        print("Sent message")
                        self.queue.task_done()
                    else:
                        # Put the message back in the queue
                        self.queue.put(message)
                        # Try to reconnect
                        self._connect()

            except Exception as e:
                print(f"Send error: {e}, reconnecting...")
                with self.lock:
                    self.connected = False
                    if self.socket:
                        try:
                            self.socket.close()
                        except:
                            pass
                # Reconnect logic
                self._connect()

    def _worker(self):
        # Ensure the connection is established first
        self._connect()

        # Start sending messages
        self._send_data()

    def __call__(self, message):
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
            # Use put_nowait to prevent blocking if queue is full
            self.queue.put_nowait(framed_message)
        except queue.Full:
            # Handle full queue - could drop or log locally
            print("Queue is full. Message not sent.")

    def close(self):
        self.shutdown = True
        # Wait for the worker thread to finish
        if hasattr(self, "worker_thread") and self.worker_thread.is_alive():
            self.worker_thread.join(timeout=1.0)
        with self.lock:
            if self.socket:
                try:
                    self.socket.close()
                except:
                    pass
        print("Connection closed.")
